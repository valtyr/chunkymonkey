chunkymonkey: build

intercept: build

build:
	./build.sh

test:
	./test.sh

docs: diagrams/top-level-architecture.png

%.png: %.dot
	dot -Tpng $< -o $@

.PHONY: build test docs
