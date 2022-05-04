#! /usr/bin/env bash

set -exuo pipefail

godot sync
/root/bin/bat -h
[ -d /root/new-bin/neovim ] || exit 1
