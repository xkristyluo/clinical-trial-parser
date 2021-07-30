#!/usr/bin/env bash
# Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved.
#
# Parse clinical-trial eligibility criteria with CFG.
#
# ./script/cfg_parse.sh

set -eu

CMD="tests/cfg/cfg.go"
CONFIG="src/resources/config/cfg.conf"

if ! go run "$CMD" -conf "$CONFIG" -logtostderr
then
  echo "CFG parser failed."
  exit 1
fi