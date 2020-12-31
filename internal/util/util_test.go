package util

import "testing"

func TestToTemplateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "leading dot", input: "~/.zshrc", want: "dot_zshrc"},
		{name: "no leading dot", input: "~/.config/nvim/init.vim", want: "init.vim"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ToTemplateName(tc.input)
			if got != tc.want {
				t.Fatalf("template incorrect, got %q want %q", got, tc.want)
			}
		})
	}
}

func TestReplaceTilde(t *testing.T) {
	getter := &OSHomeDir{}
	home, err := getter.GetHomeDir()
	if err != nil {
		t.Fatalf("error getting home dir")
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "with tilde", input: "~/.zshrc", want: home + "/.zshrc"},
		{name: "without tilde", input: "/path/to/.zshrc", want: "/path/to/.zshrc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getter.ReplaceTilde(tc.input)
			if err != nil {
				t.Fatalf("%v", err)
			}
			if got != tc.want {
				t.Fatalf("incorrect substitution, got %q want %q", got, tc.want)
			}
		})
	}
}

func TestReplaceWithTilde(t *testing.T) {
	getter := &OSHomeDir{}
	home, err := getter.GetHomeDir()
	if err != nil {
		t.Fatalf("error getting home dir")
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "with tilde", input: home + "/.zshrc", want: "~/.zshrc"},
		{name: "without tilde", input: "/path/to/.zshrc", want: "/path/to/.zshrc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getter.ReplaceWithTilde(tc.input)
			if err != nil {
				t.Fatalf("%v", err)
			}
			if got != tc.want {
				t.Fatalf("incorrect substitution, got %q want %q", got, tc.want)
			}
		})
	}
}
