package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/nokia/ntt/ttcn3"
)

func main() {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal input
	var src ttcn3.Source
	if err := json.Unmarshal(b, &src); err != nil {
		log.Fatal("ttcn3-gen-stdout: decode: ", err.Error())
	}

	// Marshal input
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(src); err != nil {
		log.Fatal("ttcn3-gen-stdout: encode: ", err.Error())
	}
}
