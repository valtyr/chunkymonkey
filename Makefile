BINARIES=\
	bin/chunkymonkey \
	bin/datatests \
	bin/inspectlevel \
	bin/intercept \
	bin/noise \
	bin/replay \
	bin/style

MOCK_FILES=\
	pkg/chunkymonkey/command/mock_icommandhandler_test.go \
	pkg/chunkymonkey/stub/mock_stub_test.go

GD_OPTS=-quiet

DIAGRAMS=diagrams/top-level-architecture.png

all: $(BINARIES)

clean:
	@-rm -f $(BINARIES)
	@gd $(GD_OPTS) -lib _obj -clean pkg
	@gd $(GD_OPTS) -lib _test -clean .

fmt:
	@gd $(GD_OPTS) -fmt -tab pkg

check: bin/style
	@bin/style `find . -name \*.go`

test: mocks
	@gd $(GD_OPTS) -lib _test -test pkg

bench: mocks
	@gd $(GD_OPTS) -lib _test -bench 'Benchmark' -match '^$$' -test pkg

libs:
	@gd $(GD_OPTS) -lib _obj pkg

test_data: bin/datatests
	@bin/datatests

bin/%: libs
	@gd $(GD_OPTS) -I _obj -lib _obj -M cmd/$*/main -output $@ pkg

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

mocks: $(MOCK_FILES)

pkg/chunkymonkey/command/mock_icommandhandler_test.go: pkg/chunkymonkey/command/icommandhandler.go
	mockgen -package command -destination $@ -source $<

pkg/chunkymonkey/stub/mock_stub_test.go: pkg/chunkymonkey/stub/stub.go
	mockgen -package stub -destination $@ -source $< -imports .=chunkymonkey/types


.PHONY: all bench check clean docs fmt mocks test test_data
