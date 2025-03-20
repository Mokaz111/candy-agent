#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)
BinaryName=candy-agent
echo "$CURDIR/bin/${BinaryName}"
exec $CURDIR/bin/${BinaryName}