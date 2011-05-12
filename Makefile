SERVER_BINARY=chunkymonkey
EXTRA_BINARIES=intercept inspectlevel datatests
DIAGRAMS=diagrams/top-level-architecture.png

server: $(SERVER_BINARY)

all: server extra

extra: $(EXTRA_BINARIES)

clean:
	@-rm $(SERVER_BINARY) $(EXTRA_BINARIES)
	@gd -q -c src
	@gd -q -c test_obj

fmt:
	@gd -q -fmt --tab src

test: datatests
	@-rm -r src/tmp*
	@-rm -r test_obj/tmp*
	@mkdir -p test_obj
	@gd -q -L test_obj -t src
	@./datatests

libs:
	@gd -q src/chunkymonkey

chunkymonkey: libs
	@gd -q -I src -o $@ src/main

intercept: libs
	@gd -q -I src -o $@ src/util/$@

inspectlevel: libs
	@gd -q -I src -o $@ src/util/$@

datatests: libs
	@gd -q -I src -o $@ src/util/$@

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

.PHONY: all clean docs extra fmt libs server test
