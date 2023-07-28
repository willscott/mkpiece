package main

import (
	"io"
	"os"

	mkpiece "github.com/willscott/mkpiece/lib"
)

func main() {
	// usage: mkpiece a.car b.car c.car ... > out.dat

	readers := make([]io.ReadSeeker, 0)
	for i, arg := range os.Args {
		if i == 0 {
			// ignore the binary itself
			continue
		}
		r, err := os.Open(arg)
		if err != nil {
			panic(err)
		}
		readers = append(readers, r)
	}
	out := mkpiece.MakeDataSegmentPiece(readers)
	io.Copy(os.Stdout, out)
}
