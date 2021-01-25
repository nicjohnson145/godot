// +build !integration

package bootstrap

import (
	"fmt"
	"github.com/nicjohnson145/godot/internal/help"
	"testing"
)

type mockItem struct {
	CheckReturn  bool
	CheckError   error
	InstallError error

	CheckCalls   int
	InstallCalls int
}

func (i *mockItem) Check() (bool, error) {
	i.CheckCalls += 1
	return i.CheckReturn, i.CheckError
}

func (i *mockItem) Install() error {
	i.InstallCalls += 1
	return i.InstallError
}

func assertCallCount(t *testing.T, method string, got int, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("%q call count incorrect, got %d want %d", method, got, want)
	}
}

func TestRunSingle(t *testing.T) {
	t.Run("already_installed", func(t *testing.T) {
		m := &mockItem{CheckReturn: true, CheckError: nil, InstallError: nil}
		r := runner{}
		err := r.RunSingle(m)
		help.Ensure(t, err)

		assertCallCount(t, "check", m.CheckCalls, 1)
		assertCallCount(t, "install", m.InstallCalls, 0)
	})

	t.Run("not_installed", func(t *testing.T) {
		m := &mockItem{CheckReturn: false, CheckError: nil, InstallError: nil}
		r := runner{}
		err := r.RunSingle(m)
		help.Ensure(t, err)

		assertCallCount(t, "check", m.CheckCalls, 1)
		assertCallCount(t, "install", m.InstallCalls, 1)
	})

	t.Run("check_error", func(t *testing.T) {
		m := &mockItem{CheckReturn: false, CheckError: fmt.Errorf("boo"), InstallError: nil}
		r := runner{}
		err := r.RunSingle(m)
		help.ShouldError(t, err)

		assertCallCount(t, "check", m.CheckCalls, 1)
		assertCallCount(t, "install", m.InstallCalls, 0)
	})
}
