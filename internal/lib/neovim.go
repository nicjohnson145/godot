package lib

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/hashicorp/go-multierror"
	"github.com/mholt/archives"
	"github.com/rs/zerolog"
)

var _ Executor = (*Neovim)(nil)

type Neovim struct {
	Name string         `yaml:"-"`
	Tag  string         `yaml:"tag" mapstructure:"tag"`
	log  zerolog.Logger `yaml:"-"`
}

func (n *Neovim) Type() ExecutorType {
	return ExecutorTypeNeovim
}

func (n *Neovim) Validate() error {
	var errs *multierror.Error

	if n.Tag == "" {
		errs = multierror.Append(errs, fmt.Errorf("tag is required"))
	}

	return errs.ErrorOrNil()
}

func (n *Neovim) SetLogger(log zerolog.Logger) {
	n.log = log
}

func (n *Neovim) GetName() string {
	return n.Name
}

func (n *Neovim) SetName(val string) {
	n.Name = val
}

func (n *Neovim) Execute(usrConf UserConfig, _ SyncOpts, _ GodotConfig) error {
	n.log.Info().Msg("ensuring neovim")

	if err := n.downloadAndUnpack(usrConf); err != nil {
		return fmt.Errorf("error downloading and unpacking: %w", err)
	}

	if err := n.symlink(usrConf); err != nil {
		return fmt.Errorf("error symlinking: %w", err)
	}

	return nil
}

func (n *Neovim) downloadAndUnpack(usrConf UserConfig) error {
	outPath, err := getDestination(usrConf, "neovim", n.Tag)
	if err != nil {
		return fmt.Errorf("error computing destination path: %w", err)
	}

	n.log.Debug().Msgf("checking if %v exists", outPath)
	exists, err := pathExists(outPath)
	if err != nil {
		return fmt.Errorf("error checking for directory existence: %w", err)
	}

	if exists {
		n.log.Debug().Msg("path already exists, nothing to download")
		return nil
	}

	n.log.Debug().Msg("not found, will download")
	gh := GithubRelease{
		Repo: "neovim/neovim",
		Tag:  n.Tag,
		AssetPatterns: map[string]map[string]string{
			"linux": {
				"amd64": "^nvim-linux64.tar.gz$",
			},
			"darwin": {
				"arm64": "^nvim-macos-arm64.tar.gz$",
			},
		},
	}

	release, err := gh.getRelease(usrConf)
	if err != nil {
		return fmt.Errorf("error getting release asset: %w", err)
	}

	dir, err := os.MkdirTemp("", "godot-neovim-")
	if err != nil {
		return fmt.Errorf("unable to make temp directory")
	}
	defer os.RemoveAll(dir)

	downloadPath := filepath.Join(dir, filepath.Base(release.DownloadUrl))
	req := requests.
		URL(release.DownloadUrl).
		ToFile(downloadPath)

	if usrConf.GithubAuth != "" {
		req.Header("Authorization", usrConf.GithubAuth)
	}
	req.Header("Accept", "application/octet-stream")

	if err := req.Fetch(context.TODO()); err != nil {
		return fmt.Errorf("error downloading from url: %w", err)
	}

	input, err := os.Open(downloadPath)
	if err != nil {
		return fmt.Errorf("error opening downloaded .tar.gz: %w", err)
	}
	defer input.Close()

	if err := os.MkdirAll(outPath, 0775); err != nil {
		return fmt.Errorf("error making final output directory: %w", err)
	}

	format := archives.CompressedArchive{
		Compression: archives.Gz{},
		Extraction:  archives.Tar{},
	}

	extractedDirName := strings.TrimSuffix(filepath.Base(release.DownloadUrl), ".tar.gz")

	err = format.Extract(context.Background(), input, func(ctx context.Context, info archives.FileInfo) error {

		infoPath := strings.TrimPrefix(info.NameInArchive, extractedDirName + "/")
		// i.e its the top level directory
		if infoPath == "" {
			return nil
		}

		dstPath := filepath.Join(outPath, infoPath)
		if info.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return fmt.Errorf("error replicating directory form archive: %w", err)
			}
			return nil
		}

		fl, err := info.Open()
		if err != nil {
			return fmt.Errorf("error opening file in archive: %w", err)
		}
		defer fl.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("error creating destination file: %w", err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, fl); err != nil {
			return fmt.Errorf("error copying file: %w", err)
		}

		if err := os.Chmod(dstPath, info.Mode()); err != nil {
			return fmt.Errorf("error chmod-ing file: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error extracting archive: %w", err)
	}

	return nil
}

func (n *Neovim) symlink(usrConf UserConfig) error {
	outPath, err := getDestination(usrConf, "neovim", n.Tag)
	if err != nil {
		return fmt.Errorf("error computing destination path: %w", err)
	}

	symlinkName := filepath.Join(filepath.Dir(outPath), "neovim")
	if err := createSymlink(outPath, symlinkName); err != nil {
		return fmt.Errorf("error creating symlink: %w", err)
	}

	return nil
}
