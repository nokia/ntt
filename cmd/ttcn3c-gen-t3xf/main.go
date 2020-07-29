package main

import (
	"bytes"
	"log"
	"os"

	"github.com/nokia/ntt/plugin"
	"google.golang.org/protobuf/proto"
)

func main() {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(os.Stdin); err != nil {
		log.Fatal(err)
	}
	req := &plugin.GeneratorRequest{}
	if err := proto.Unmarshal(buf.Bytes(), req); err != nil {
		log.Fatal(err)
	}

}
