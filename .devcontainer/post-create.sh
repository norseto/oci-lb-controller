#!/usr/bin/env bash

sudo chown -R $(id -u):$(id -g) \
  /home/vscode/.aws /home/vscode/.kube /home/vscode/.cache \
  /usr/local/go \
  /tmp/.gocache /tmp/.gomodcache /go

sudo chown -R $(id -u):$(id -g) $HOME/.codex $HOME/.claude
