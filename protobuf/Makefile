.PHONY: generate
generate: generate-go generate-cpp generate-python

.PHONY: generate-go
generate-go:
	protoc --go_out=. --go_opt=paths=source_relative *.proto

.PHONY: generate-cpp
generate-cpp:
	protoc --cpp_out=. *.proto

.PHONY: generate-python
generate-python:
	protoc --python_out=. *.proto
