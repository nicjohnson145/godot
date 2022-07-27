package lib

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReplaceTilde(t *testing.T) {
	testData := []struct {
		name string
		str  string
		want string
	}{
		{
			name: "with_tilde",
			str:  "~/bin/some-bin",
			want: "/home/njohnson/bin/some-bin",
		},
		{
			name: "no_tilde",
			str:  "/bin/some-bin",
			want: "/bin/some-bin",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			got := replaceTilde(tc.str, "/home/njohnson")
			require.Equal(t, tc.want, got)
		})
	}
}

func TestTagNormalization(t *testing.T) {
	testData := []struct {
		name string
		tag  string
		want string
	}{
		{
			name: "kustomize",
			tag:  "kustomize/v4.5.4",
			want: "kustomize-v4.5.4",
		},
		{
			name: "regular",
			tag:  "v0.21.0",
			want: "v0.21.0",
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, normalizeTag(tc.tag))
		})
	}
}
