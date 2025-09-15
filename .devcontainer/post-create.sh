#!/usr/bin/env bash

sudo chown -R vscode:vscode \
  /home/vscode/.aws /home/vscode/.kube /home/vscode/.cache \
  /usr/local/go /go/bin /go/pkg /home/vscode/.gocache
