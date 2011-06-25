SERVER_BINARY=chunkymonkey
EXTRA_BINARIES=\
	datatests \
	inspectlevel \
	intercept \
	noise \
	replay \
	style

GD_OPTS=-quiet -lib _obj src

DIAGRAMS=diagrams/top-level-architecture.png

server: $(SERVER_BINARY)

all: server extra

extra: $(EXTRA_BINARIES)

cleanobj:
	@gd $(GD_OPTS) -clean

clean: cleanobj
	@-rm -f $(SERVER_BINARY) $(EXTRA_BINARIES)

fmt:
	@gd $(GD_OPTS) -fmt --tab

check: style
	@./style `find . -name \*.go`

# requires clean-up due to bug in godag
test: cleanobj
	@gd $(GD_OPTS) -test

# Note that this will also compile code in the src/util directory.
libs:
	@gd $(GD_OPTS)

test_data: datatests
	@./datatests

bench:
	@gd $(GD_OPTS) -bench . -match "Regex That Matches 0 Tests" -test

chunkymonkey: libs
	@gd $(GD_OPTS) -output $@ -main ^main$$

datatests: libs
	@gd $(GD_OPTS) -output $@ -main $@

intercept: libs
	@gd $(GD_OPTS) -output $@ -main $@

inspectlevel: libs
	@gd $(GD_OPTS) -output $@ -main $@

noise: libs
	@gd $(GD_OPTS) -output $@ -main $@

replay: libs
	@gd $(GD_OPTS) -output $@ -main $@

style: src/util/style/style.go
	@gd $(GD_OPTS) -output $@ -main $@

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

.PHONY: all bench check clean cleanobj docs extra fmt server test test_data
