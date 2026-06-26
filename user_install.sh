#!/usr/bin/env bash

SCRIPTLOCATION=$(dirname -- "$(readlink -f -- "$BASH_SOURCE")")
cd $SCRIPTLOCATION

echo "compiling..."
go build wotd.go
echo "copying to $HOME/.local/bin ..."
cp wotd $HOME/.local/bin
echo "done!"
