package main

import (
	"fmt"
	"io"
	"math/bits"
	"os"

	"github.com/filecoin-project/go-data-segment/datasegment"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/filecoin-project/go-state-types/abi"
)

func main() {
	// usage: mkpiece a.car b.car c.car ... > out.dat

	readers := make([]io.Reader, 0)
	deals := make([]abi.PieceInfo, 0)
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
		cp := new(commp.Calc)
		io.Copy(cp, r)
		rawCommP, size, err := cp.Digest()
		if err != nil {
			panic(err)
		}
		r.Seek(0, io.SeekStart)
		c, _ := commcid.DataCommitmentV1ToCID(rawCommP)
		subdeal := abi.PieceInfo{
			Size:     abi.PaddedPieceSize(size),
			PieceCID: c,
		}
		deals = append(deals, subdeal)
	}
	if len(deals) == 0 {
		fmt.Printf("Usage: mkpiece <a.car> ... > out.dat\r\n")
		return
	}

	_, size, err := datasegment.ComputeDealPlacement(deals)
	if err != nil {
		panic(err)
	}

	overallSize := abi.PaddedPieceSize(size)
	// we need to make this the 'next' power of 2 in order to have space for the index
	next := 1 << (64 - bits.LeadingZeros64(uint64(overallSize)))

	a, err := datasegment.NewAggregate(abi.PaddedPieceSize(next), deals)
	if err != nil {
		panic(err)
	}
	out, err := a.AggregateObjectReader(readers)
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)
}
