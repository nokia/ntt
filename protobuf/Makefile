.PHONY: generate
generate: generate-go

.PHONY: generate-go
generate-go:
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative *.proto

.PHONY: generate-cpp
generate-cpp:
	protoc --cpp_out=. *.proto

.PHONY: generate-python
generate-python:
	protoc --python_out=. *.proto
