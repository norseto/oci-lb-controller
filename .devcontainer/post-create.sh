#!/usr/bin/env bash

sudo chown -R $(id -u):$(id -g) \
  /home/vscode/.aws /home/vscode/.kube /home/vscode/.cache \
  ${HOME}/.codex \
  ${HOME}/.cursor \
  /usr/local/go \
  /tmp/.gocache /tmp/.gomodcache /go
