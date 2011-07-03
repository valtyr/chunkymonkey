BINARIES=\
	bin/chunkymonkey \
	bin/datatests \
	bin/inspectlevel \
	bin/intercept \
	bin/noise \
	bin/replay \
	bin/style

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

test:
	@gd $(GD_OPTS) -lib _test -test pkg

bench:
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

.PHONY: all bench check clean docs fmt test test_data
