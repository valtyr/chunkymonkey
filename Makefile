SERVER=chunkymonkey
BINARIES=intercept inspectlevel dumpblockdefs
DIAGRAMS=diagrams/top-level-architecture.png

all: $(SERVER)

extra: $(BINARIES)

clean:
	@gd -q -clean src

fmt:
	@gd -q -fmt src

test:
	@mkdir -p .test_obj
	@gd -q -L .test_obj -t src/lib

libs:
	@gd -q src/lib

chunkymonkey: libs
	@gd -q -I src/lib -o $@ src/$@

intercept: libs
	@gd -q -I src/lib -o $@ src/$@

inspectlevel: libs
	@gd -q -I src/lib -o $@ src/$@

dumpblockdefs: libs
	@gd -q -I src/lib -o $@ src/$@

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

.PHONY: all clean docs extra fmt libs test
