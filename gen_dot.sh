#!/bin/sh

# Script that generates a diagram via GoDag and Dot to produce a dependency
# tree diagram for the project packages. Dependencies upon packages not in the
# project are omitted for clarity.

gd -dot deps.dot src \
    && sed -ri '/->/{/"[^"]+" -> "(cmd|chunkymonkey|nbt|perlin|testencoding|testmatcher)[/"]/b ok ; d ; : ok}' deps.dot \
    && dot -O -Tpng deps.dot \
    && eog deps.dot.png

