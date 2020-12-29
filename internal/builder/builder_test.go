package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func setupBuilder(t *testing.T, target string) (string, string, func()) {
	home, remove := help.CreateTempDir(t, "home")

	dotPath := filepath.Join(home, "dotfiles")
	err := os.Mkdir(dotPath, 0744)
	if err != nil {
		t.Errorf("error making dir, %v", err)
	}

	err = os.Mkdir(filepath.Join(dotPath, "templates"), 0744)
	if err != nil {
		t.Errorf("error making dir, %v", err)
	}
	help.WriteConfig(t, home, fmt.Sprintf(`{"target": "%v", "dotfiles_root": "%v"}`, target, dotPath))

	return home, dotPath, remove
}

func setupBasicConfig(t *testing.T, dotPath string, data *string) {
	t.Helper()

	help.WriteRepoConf(t, dotPath, `{
		"all_files": {
			"dot_zshrc": "~/.zshrc"
		},
		"renders": {
			"host": ["dot_zshrc"]
		}
	}`)

	if data != nil {
		help.WriteData(t, filepath.Join(dotPath, "templates", "dot_zshrc"), *data)
	}
}

func TestBuilder(t *testing.T) {
	t.Run("simple no-template integration", func(t *testing.T) {
		home, dotPath, remove := setupBuilder(t, "host")
		defer remove()

		expected := "zsh contents"
		setupBasicConfig(t, dotPath, &expected)
		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("templated integration", func(t *testing.T) {
		home, dotPath, remove := setupBuilder(t, "host")
		defer remove()

		expected := "host zsh contents"
		template := `{{ if oneOf . "host" "other" }}{{ .Target }}{{ end }} zsh contents`
		setupBasicConfig(t, dotPath, &template)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("file exists as symlink", func(t *testing.T) {
		home, dotPath, remove := setupBuilder(t, "host")
		defer remove()

		expected := "host zsh contents"
		setupBasicConfig(t, dotPath, &expected)
		err := os.Mkdir(filepath.Join(dotPath, "build"), 0744)
		if err != nil {
			t.Fatalf("err creating dir, %v", err)
		}
		buildFile := filepath.Join(dotPath, "build/dot_zshrc")
		help.WriteData(t, buildFile, "orig")
		os.Symlink(filepath.Join(home, ".zshrc"), filepath.Join(dotPath, "build/dot_zshrc"))

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build(false)
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("file exists as file no force", func(t *testing.T) {
		home, dotPath, remove := setupBuilder(t, "host")
		defer remove()

		destPath := filepath.Join(home, ".zshrc")
		help.WriteData(t, destPath, "")

		expected := "zsh contents"
		setupBasicConfig(t, dotPath, &expected)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(false)
		if err == nil {
			t.Fatalf("code should have errored")
		}

		help.AssertFileContents(t, destPath, "")
	})

	t.Run("file exists as file with force", func(t *testing.T) {
		home, dotPath, remove := setupBuilder(t, "host")
		defer remove()

		destPath := filepath.Join(home, ".zshrc")
		help.WriteData(t, destPath, "")

		expected := "zsh contents"
		setupBasicConfig(t, dotPath, &expected)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build(true)
		if err != nil {
			t.Fatalf("error building, %v", err)
		}

		help.AssertFileContents(t, destPath, "zsh contents")
	})
}
