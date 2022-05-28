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
