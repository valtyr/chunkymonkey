#!/bin/bash
. common.sh

clean_for "build"

BINARIES="chunkymonkey intercept inspectlevel"

gd src/lib || exit $?

for BINARY in $BINARIES; do
    gd -o $BINARY -I src/lib src/$BINARY || exit $?
done
