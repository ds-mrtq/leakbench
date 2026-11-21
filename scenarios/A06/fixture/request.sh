#!/usr/bin/env bash
set -eu
curl --fail-with-body -H 'Authorization: Bearer retired' http://127.0.0.1:8091/status
