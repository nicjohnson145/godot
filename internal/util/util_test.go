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

func TestReplacePrefix(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		prefix    string
		newPrefix string
		want      string
	}{
		{name: "replace tilde exists", path: "~/.zshrc", prefix: "~/", newPrefix: "/home/dir", want: "/home/dir/.zshrc"},
		{name: "replace tilde missing", path: "/path/to/.zshrc", prefix: "~/", newPrefix: "/home/dir", want: "/path/to/.zshrc"},
		{name: "replace with tilde exists", path: "/home/dir/.zshrc", prefix: "/home/dir", newPrefix: "~/", want: "~/.zshrc"},
		{name: "replace with tilde missing", path: "/path/to/.zshrc", prefix: "/home/dir", newPrefix: "~/", want: "/path/to/.zshrc"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ReplacePrefix(tc.path, tc.prefix, tc.newPrefix)
			if got != tc.want {
				t.Fatalf("replacement incorrect, got %q want %q", got, tc.want)
			}
		})
	}
}
