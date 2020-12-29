package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicjohnson145/godot/internal/help"
)

func TestBuilder(t *testing.T) {
	t.Run("simple no-template integration", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotPath := filepath.Join(home, "dotfiles")
		err := os.Mkdir(dotPath, 0744)
		if err != nil {
			t.Errorf("error making dir, %v", err)
		}

		err = os.Mkdir(filepath.Join(dotPath, "templates"), 0744)
		if err != nil {
			t.Errorf("error making dir, %v", err)
		}
		expected := "zsh contents"
		help.WriteData(t, filepath.Join(dotPath, "templates", "dot_zshrc"), expected)
		help.WriteConfig(t, home, fmt.Sprintf(`{"target": "host", "dotfiles_root": "%v"}`, dotPath))
		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc"
			},
			"renders": {
				"host": ["dot_zshrc"]
			}
		}`)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build()
		if err != nil {
			t.Errorf("error encountered, %v", err)
		}

		bytes, err := ioutil.ReadFile(filepath.Join(home, ".zshrc"))
		if err != nil {
			t.Errorf("error reading, %v", err)
		}

		content := string(bytes)
		if content != expected {
			t.Errorf("unexpected contents, got %q want %q", content, expected)
		}
	})

	t.Run("templated integration", func(t *testing.T) {
		home, remove := help.CreateTempDir(t, "home")
		defer remove()

		dotPath := filepath.Join(home, "dotfiles")
		err := os.Mkdir(dotPath, 0744)
		if err != nil {
			t.Errorf("error making dir, %v", err)
		}

		err = os.Mkdir(filepath.Join(dotPath, "templates"), 0744)
		if err != nil {
			t.Errorf("error making dir, %v", err)
		}
		expected := "host zsh contents"
		help.WriteData(t, filepath.Join(dotPath, "templates", "dot_zshrc"), `{{ .Target }} zsh contents`)
		help.WriteConfig(t, home, fmt.Sprintf(`{"target": "host", "dotfiles_root": "%v"}`, dotPath))
		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc"
			},
			"renders": {
				"host": ["dot_zshrc"]
			}
		}`)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err = b.Build()
		if err != nil {
			t.Errorf("error encountered, %v", err)
		}

		bytes, err := ioutil.ReadFile(filepath.Join(home, ".zshrc"))
		if err != nil {
			t.Errorf("error reading, %v", err)
		}

		content := string(bytes)
		if content != expected {
			t.Errorf("unexpected contents, got %q want %q", content, expected)
		}
	})
}
