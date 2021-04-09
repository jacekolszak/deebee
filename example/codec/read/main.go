package main

import (
	"fmt"
	"io/ioutil"

	"github.com/jacekolszak/deebee/codec"
	"github.com/jacekolszak/deebee/store"
)

// This example shows primitive decoder reading all bytes into memory
func main() {
	s, err := store.Open("/tmp/deebee")
	if err != nil {
		panic(err)
	}

	var bytesRead []byte
	version, err := codec.Read(s, func(reader store.Reader) error {
		all, e := ioutil.ReadAll(reader)
		if e != nil {
			return e
		}
		bytesRead = all
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Bytes read: %s\n", bytesRead)
	fmt.Printf("Version %+v", version)
}
