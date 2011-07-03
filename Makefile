BINARIES=\
	bin/chunkymonkey \
	bin/datatests \
	bin/inspectlevel \
	bin/intercept \
	bin/noise \
	bin/replay \
	bin/style

MOCK_FILES=\
	src/chunkymonkey/command/mock_icommandhandler_test.go \
	src/chunkymonkey/stub/mock_stub_test.go

GD_OPTS=-quiet

DIAGRAMS=diagrams/top-level-architecture.png

all: $(BINARIES)

clean:
	@-rm -f $(BINARIES)
	@gd $(GD_OPTS) -lib _obj -clean src
	@gd $(GD_OPTS) -lib _test -clean .

fmt:
	@gd $(GD_OPTS) -fmt -tab src

check: bin/style
	@bin/style `find . -name \*.go`

test: mocks
	@gd $(GD_OPTS) -lib _test -test src

bench: mocks
	@gd $(GD_OPTS) -lib _test -bench 'Benchmark' -match '^$$' -test src

libs:
	@gd $(GD_OPTS) -lib _obj src

test_data: bin/datatests
	@bin/datatests

bin/%: libs
	@gd $(GD_OPTS) -I _obj -lib _obj -M cmd/$*/main -output $@ src

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

mocks: $(MOCK_FILES)

src/chunkymonkey/command/mock_icommandhandler_test.go: src/chunkymonkey/command/icommandhandler.go
	mockgen -package command -destination $@ -source $<

src/chunkymonkey/stub/mock_stub_test.go: src/chunkymonkey/stub/stub.go
	mockgen -package stub -destination $@ -source $< -imports .=chunkymonkey/types


.PHONY: all bench check clean docs fmt mocks test test_data
