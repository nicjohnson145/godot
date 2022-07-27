#! /usr/bin/env bash

set -exuo pipefail

godot sync -v

HOMEDIR="/home/newuser"

$HOMEDIR/bin/bat -h
$HOMEDIR/bin/kustomize -h

[ -d $HOMEDIR/new-bin/neovim ] || exit 1

[ -f $HOMEDIR/new-bin/conf-dir/test.config ] || exit 1
ACTUAL_CONF=$(cat $HOMEDIR/new-bin/conf-dir/test.config)
if [[ "$ACTUAL_CONF" != "Hello from test" ]]; then
    echo "Got bad config"
    echo $ACTUAL_CONF
    exit 1
fi

# Tmux should be installed through apt, `tmux --help` has an exit code of 1, so just make sure it's
# in the path
which tmux

# Kubectl should be installed
$HOMEDIR/bin/kubectl -h

# So should vault
$HOMEDIR/bin/vault -h
