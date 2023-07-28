package main

import (
	"fmt"
	"io"
	"os"

	"github.com/filecoin-project/go-data-segment/datasegment"
	"github.com/filecoin-project/go-state-types/abi"
)

func main() {
	// usage: splitpiece out.dat folder

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	fi, err := file.Stat()
	if err != nil {
		panic(err)
	}

	offset := datasegment.DataSegmentIndexStartOffset(abi.UnpaddedPieceSize(fi.Size()).Padded())
	file.Seek(int64(offset), io.SeekStart)
	index, err := datasegment.ParseDataSegmentIndex(file)
	if err != nil {
		panic(err)
	}
	entries, err := index.ValidEntries()
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		file.Seek(0, io.SeekStart)
		strt := e.UnpaddedOffest()
		leng := e.UnpaddedLength()
		segment := io.NewSectionReader(file, int64(strt), int64(leng))
		seg, err := os.OpenFile(fmt.Sprintf("%s/%d.chunk", os.Args[2], strt), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		n, err := io.Copy(seg, segment)
		if err != nil {
			panic(err)
		}
		if n != int64(leng) {
			panic("didn't write enough")
		}
		seg.Close()
		fmt.Printf("Segment found: %d - %d\n", strt, strt+leng)
	}
}
