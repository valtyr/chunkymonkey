#!/bin/bash
. common.sh

clean_for "test"

gd -t src/lib $*
