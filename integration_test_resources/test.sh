#! /usr/bin/env bash

set -exuo pipefail

godot sync -v

HOMEDIR="/home/newuser"

# devtools should exist
$HOMEDIR/bin/devtools -h
# Bat should exist
$HOMEDIR/bin/bat -h
# Kubectl should be installed
$HOMEDIR/bin/kubectl -h
# Fd should not be installed
if [[ -f $HOMEDIR/bin/fd ]]; then
    echo "fd should not be installed"
    exit 1
fi
# zsh should be installed from apt
which zsh

# pyenv should be check out to a specific commit
PYENV_SHA="304515f2cdd11db151b7a5733d11934d3990a67e"
if [[ "$(git -C $HOMEDIR/.pyenv rev-parse HEAD)" != "$PYENV_SHA" ]]; then
    echo "pyenv should be checked out to $PYENV_SHA"
    exit 1
fi

# Foo config should be rendered with this content
FOO_CONFIG_PATH="$HOMEDIR/.config/foo_config/config.txt"
[ -f $FOO_CONFIG_PATH ] || exit 1
grep "bundle-o-things installed" $FOO_CONFIG_PATH
grep "kubectl installed" $FOO_CONFIG_PATH
if grep "fd installed" $FOO_CONFIG_PATH; then
        echo "fd should not show as installed"
        exit 1
fi

# Golang should be installed at 1.19.3
export PATH=$PATH:/usr/local/go/bin
if [[ "$(go version)" != "go version go1.19.3 linux/amd64" ]]; then
    echo "correct go version not installed"
    exit 1
fi

# Gopls should be installed
$HOMEDIR/go/bin/gopls -h


exit 0
