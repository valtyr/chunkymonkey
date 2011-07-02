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

clean: cleantestobj
	@-rm -f $(BINARIES)
	@gd $(GD_OPTS) -lib _obj -clean .

fmt:
	@gd $(GD_OPTS) -fmt -tab pkg
	@gd $(GD_OPTS) -fmt -tab cmd

check: bin/style
	@bin/style `find . -name \*.go`

cleantestobj:
	@gd $(GD_OPTS) -lib _test -clean .

# requires clean-up due to bug in godag
test: cleantestobj
	@gd $(GD_OPTS) -lib _test/pkg -test pkg

bench: cleantestobj
	@gd $(GD_OPTS) -lib _test/pkg -bench 'Bench' -match 'Regex That Matches 0 Tests' -test pkg

libs:
	@gd $(GD_OPTS) -lib _obj/pkg pkg

test_data: bin/datatests
	@bin/datatests

bin/%: libs
	@gd $(GD_OPTS) -I _obj/pkg -output $@ cmd/$*

docs: $(DIAGRAMS)

%.png: %.dot
	@dot -Tpng $< -o $@

.PHONY: all bench check clean cleantestobj docs fmt test test_data
