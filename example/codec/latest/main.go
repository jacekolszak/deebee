package main

import (
	"fmt"

	"github.com/jacekolszak/deebee/codec"
	"github.com/jacekolszak/deebee/json"
	"github.com/jacekolszak/deebee/store"
)

// This example shows how to read a latest version, which is not corrupted and deserializable, and fail-over
// to previous version if so
func main() {
	s, err := store.Open("/tmp/deebee/json")
	if err != nil {
		panic(err)
	}

	out := &State{}
	version, err := codec.ReadLatest(s, json.Decoder(out))
	if err != nil {
		panic(err)
	}

	fmt.Printf("State read: %+v\n", out)
	fmt.Printf("Version %+v", version)
}

type State struct {
	Name string
	Age  int
}
