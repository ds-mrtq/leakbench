#!/usr/bin/env bash
set -eu
git config --local remote.origin.pushurl "$HOME/remote/retired.git"
git push origin main
