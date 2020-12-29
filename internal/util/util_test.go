package util

import "testing"

func TestToTemplateName(t *testing.T) {
	tests := []struct{
		name string
		input string
		want string
	}{
		{name: "leading dot", input: "~/.zshrc", want: "dot_zshrc"},
		{name: "no leading dot", input: "~/.config/nvim/init.vim", want: "init.vim"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T){
			got := ToTemplateName(tc.input)
			if got != tc.want {
				t.Fatalf("template incorrect, got %q want %q", got, tc.want)
			}
		})
	}
}
