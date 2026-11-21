#!/usr/bin/env bash
set -eu
: "${DEPLOY_TOKEN:?source ~/.profile first}"
mkdir -p deploy-out
cp build/package.txt deploy-out/package.txt
printf "Local deploy complete\n"
