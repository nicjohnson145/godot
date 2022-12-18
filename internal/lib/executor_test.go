package lib

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestApplyOrdering(t *testing.T) {
	inp := []Executor{
		&ConfigFile{
			Name: "cf1",
		},
		&GoInstall{
			Name: "go-inst-1",
		},
		&Golang{
			Name: "golang",
		},
	}
	got := applyOrdering(inp)
	require.Equal(
		t,
		[]string{"cf1", "golang", "go-inst-1"},
		lo.Map(got, func(e Executor, _ int) string {
			return e.GetName()
		}),
	)
}
