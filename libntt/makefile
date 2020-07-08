CC=gcc
GOBUILD=go build
CFLAGS=-shared -fPIC
GOFLAGS=-buildmode=c-archive

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
	echo "install"


.PHONY: clean
clean:
	rm $(LIB)
	rm *.a
	rm libntt.h
	
