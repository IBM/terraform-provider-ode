#!/bin/bash
# Copyright (c) IBM Corporation
# SPDX-License-Identifier: Apache-2.0



op run --env-file="$(pwd)/.vscode/.env" -- dlv $@
