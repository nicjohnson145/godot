#! /usr/bin/env bash

set -exuo pipefail

godot sync

/home/root/bin/bat -h

[ -d /root/new-bin/neovim ] || exit 1

[ -f /root/new-bin/conf-dir/test.config ] || exit 1
ACTUAL_CONF=$(cat /root/new-bin/conf-dir/test.config)
if [[ "$ACTUAL_CONF" != "Hello from test" ]]; then
    echo "Got bad config"
    echo $ACTUAL_CONF
    exit 1
fi
