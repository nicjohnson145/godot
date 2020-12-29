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

func TestBuilder(t *testing.T) {
	t.Run("simple no-template integration", func(t *testing.T) {
		home, dotPath, remove := setupBuilder(t, "host")
		defer remove()

		expected := "zsh contents"
		help.WriteData(t, filepath.Join(dotPath, "templates", "dot_zshrc"), expected)
		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc"
			},
			"renders": {
				"host": ["dot_zshrc"]
			}
		}`)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build()
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})

	t.Run("templated integration", func(t *testing.T) {
		home, dotPath, remove := setupBuilder(t, "host")
		defer remove()

		expected := "host zsh contents"
		help.WriteData(
			t,
			filepath.Join(dotPath, "templates", "dot_zshrc"),
			`{{ if oneOf .Target "host" "other" }}{{ .Target }}{{ end }} zsh contents`,
		)
		help.WriteRepoConf(t, dotPath, `{
			"all_files": {
				"dot_zshrc": "~/.zshrc"
			},
			"renders": {
				"host": ["dot_zshrc"]
			}
		}`)

		b := Builder{Getter: &help.TempHomeDir{HomeDir: home}}
		err := b.Build()
		if err != nil {
			t.Fatalf("error encountered, %v", err)
		}

		help.AssertFileContents(t, filepath.Join(home, ".zshrc"), expected)
	})
}
