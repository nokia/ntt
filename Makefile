PROJECT := github.com/nokia/ntt

# Common GNU installation folders
DESTDIR ?=
PREFIX     ?= /ust/local
BINDIR     ?= ${PREFIX}/bin
LIBEXECDIR ?= ${PREFIX}/libexec
MANDIR     ?= ${PREFIX}/share/man
ETCDIR     ?= /etc

BASHINSTALLDIR=${PREFIX}/share/bash-completion/completions
ZSHINSTALLDIR=${PREFIX}/share/zsh/site-functions

# Commands
GO   ?= go

# Go specific stuff
export GO111MODULE=off
export GOPROXY=https://proxy.golang.org

GO_BUILD=$(GO) build

# If GOPATH not specified, use one in the local directory
ifeq ($(GOPATH),)
export GOPATH := $(CURDIR)/_output
unexport GOBIN
endif
FIRST_GOPATH := $(firstword $(subst :, ,$(GOPATH)))
GOPKGDIR := $(FIRST_GOPATH)/src/$(PROJECT)
GOPKGBASEDIR ?= $(shell dirname "$(GOPKGDIR)")

GOBIN := $(shell $(GO) env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(FIRST_GOPATH)/bin
endif


# Go module support: set `-mod=vendor` to force use the vendored sources
ifeq ($(shell $(GO) help mod >/dev/null 2>&1 && echo true), true)
	GO_BUILD=GO111MODULE=on $(GO) build -mod=vendor
endif


#GCFLAGS ?= all=-trimpath=${PWD}
#ASMFLAGS ?= all=-trimpath=${PWD}
#LDFLAGS_NTT ?= \
#	  -X $(CMD_NTT)/define.gitCommit=$(GIT_COMMIT) \
#	  -X $(CMD_NTT)/define.buildInfo=$(BUILD_INFO) \
#	  -X $(CMD_NTT)/config._installPrefix=$(PREFIX) \
#	  -X $(CMD_NTT)/config._etcDir=$(ETCDIR) \
#	  $(EXTRA_LDFLAGS)
#
#SOURCES = $(shell find . -path './.*' -prune -o -name "*.go")

#COMMIT_NO ?= $(shell git rev-parse HEAD 2> /dev/null || true)
#GIT_COMMIT ?= $(if $(shell git status --porcelain --untracked-files=no),${COMMIT_NO}-dirty,${COMMIT_NO})
#DATE_FMT = %s
#ifdef SOURCE_DATE_EPOCH
#	BUILD_INFO ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "+$(DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "+$(DATE_FMT)" 2>/dev/null || date -u "+$(DATE_FMT)")
#	ISODATE ?= $(shell date -d "@$(SOURCE_DATE_EPOCH)" --iso-8601)
#else
#	BUILD_INFO ?= $(shell date "+$(DATE_FMT)")
#	ISODATE ?= $(shell date --iso-8601)
#endif

.PHONY: all ## build whole project (default)
all: build

# It's common practice to provide a help target describing available targets.
# This script greps through the Makefile and displays all phony-targets with a
# `##` help string.
.PHONY: help
help:
	@echo Available targets:
	@perl -ne 'printf("\t%-10s\t%s\n", $$1, $$2)  if /^\.PHONY:\s*(.*)\s*##\s*(.*)$$/' <$(MAKEFILE_LIST)

.PHONY: build ## build everything
build: ntt

.PHONY: check ## run tests
check:
	$(GO) test ./...

.PHONY: install ## install NTT
install: build
	install -d -m 755 $(DESTDIR)$(BINDIR)
	install -m 755 bin/ntt $(DESTDIR)$(BINDIR)/ntt


# Binaries must be PHONY-targets, because
.PHONY: ntt ## build ntt CLI
ntt:
	$(GO_BUILD) -o $@ ./cmd/ntt

.PHONY: generate ## run code generators
generate:
	$(GO) generate -x ./...

