PYTHON += /usr/bin/python3
FLAGS += build_ext --inplace
 
all: ntt.so
 
install: ntt.so
	cp libntt.so /usr/lib 
 
ntt.so:
	$(PYTHON) setup.py ${FLAGS}
 
 
clean:
	rm ntt*.so
	rm -r build

