UNAME=$(shell uname -s)
BLDDATE=$(shell date -u +%Y%m%dT%H%M%S)
VERSION?=$(shell git describe --tags `git rev-list --tags --max-count=1`)
REVISION=$(shell git rev-list --all --max-count=1 --abbrev-commit)
ARCH?=amd64
OS?=darwin linux
LDFLAGS=" -s -X main.BuildDate='${BLDDATE}' -X main.Version='${VERSION}' -X main.Revision='${REVISION}'"
TRY6D_SRCS = $(wildcard cmd/**/*.go)

APPS = try6d
BLDDIR = build

ifeq ($(UNAME),Darwin)
ECHO=echo
else
ECHO=echo -e
endif

all: $(APPS)

$(BLDDIR)/%:
	@mkdir -p $(dir $@)
	@$(ECHO) "==> Built $@ binaries"
		$(foreach os,$(OS), \
			$(foreach arch,$(ARCH), \
				$(shell GO15VENDOREXPERIMENT=1 CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -v -ldflags ${LDFLAGS} -a -installsuffix cgo -o $(abspath $@)_${os}_${arch}.bin ./$*) \
			) \
		)

$(BINARIES): %: $(BLDDIR)/%

$(APPS): %: $(BLDDIR)/cmd/%

$(BLDDIR)/cmd/try6d: $(TRY6D_SRCS)

vendor:
	@${ECHO} "==> Vendoring dependencies"
	GO15VENDOREXPERIMENT=1 godep save ./...

image: $(BINARIES)
	@docker build -t datflow/try6d:${VERSION} .

clean:
	@${ECHO} "Deleting binaries ${VERSION}"
	@rm -rfv ./build

.PHONY: clean $(BINARIES) $(APPS)
