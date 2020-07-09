CC=gcc
GOBUILD=go build
CFLAGS=-shared -fPIC
GOFLAGS=-buildmode=c-archive
DESTDIR    ?=
PREFIX     ?= /usr/local
INCLUDEDIR ?= $(PREFIX)/include
LIBDIR     ?= $(PREFIX)/lib

OS:=$(shell uname)

ifeq ($(OS), Darwin)
	LIB+=libntt.dylib
endif
ifeq ($(OS), Linux)
	LIB+=libntt.so
endif

all: $(LIB)

$(LIB): ntt.c libntt.a
	$(CC) $(CFLAGS) -o $@ $^

libntt.a: *.go
	$(GOBUILD)  -o $@ $(GOFLAGS) $^


install: all
	install -d -m 755 $(DESTDIR)$(INCLUDEDIR)
	install -d -m 755 $(DESTDIR)$(LIBDIR)
	install -m 644 $(LIB) $(DESTDIR)$(LIBDIR)
	install -m 644 ntt.h  $(DESTDIR)$(INCLUDEDIR)


.PHONY: clean
clean:
	rm $(LIB)
	rm *.a
	rm libntt.h
	
