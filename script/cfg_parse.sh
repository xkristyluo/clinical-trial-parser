#!/usr/bin/env bash
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.
#
# Parse clinical-trial eligibility criteria with CFG.
#
# ./script/cfg_parse.sh

set -eu

CMD="tests/cfg/cfg.go"
export CONFIG_FILE="src/resources/config/cfg.conf"

if ! go run "$CMD" -logtostderr
then
  echo "CFG parser failed."
  exit 1
fi