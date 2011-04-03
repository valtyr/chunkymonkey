#!/bin/bash
. common.sh

clean_for "build"

gd src/lib \
    && gd -o chunkymonkey -I src/lib src/chunkymonkey \
    && gd -o intercept -I src/lib src/intercept
