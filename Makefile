BINARIES=chunkymonkey intercept inspectlevel
DIAGRAMS=diagrams/top-level-architecture.png

all: $(BINARIES)

libs:
	@gd -q src/lib

test:
	@mkdir -p .test_obj
	@gd -q -L .test_obj -t src/lib

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

chunkymonkey: libs
	@gd -q -I src/lib -o $@ src/$@

intercept: libs
	@gd -q -I src/lib -o $@ src/$@

inspectlevel: libs
	@gd -q -I src/lib -o $@ src/$@

.PHONY: all libs test docs
