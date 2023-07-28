package mkpiece

import (
	"bytes"
	"io"
	"math/bits"

	"github.com/filecoin-project/go-data-segment/datasegment"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/filecoin-project/go-state-types/abi"
)

func MakeDataSegmentPiece(subPieces []io.ReadSeeker) io.Reader {
	readers := make([]io.Reader, 0)
	deals := make([]abi.PieceInfo, 0)
	for _, arg := range subPieces {
		readers = append(readers, arg)
		cp := new(commp.Calc)
		io.Copy(cp, arg)
		rawCommP, size, err := cp.Digest()
		if err != nil {
			panic(err)
		}
		arg.Seek(0, io.SeekStart)
		c, _ := commcid.DataCommitmentV1ToCID(rawCommP)
		subdeal := abi.PieceInfo{
			Size:     abi.PaddedPieceSize(size),
			PieceCID: c,
		}
		deals = append(deals, subdeal)
	}
	if len(deals) == 0 {
		return nil
	}

	_, size, err := datasegment.ComputeDealPlacement(deals)
	if err != nil {
		panic(err)
	}

	overallSize := abi.PaddedPieceSize(size)
	// we need to make this the 'next' power of 2 in order to have space for the index
	next := 1 << (64 - bits.LeadingZeros64(uint64(overallSize+256)))

	a, err := datasegment.NewAggregate(abi.PaddedPieceSize(next), deals)
	if err != nil {
		panic(err)
	}
	out, err := a.AggregateObjectReader(readers)
	if err != nil {
		panic(err)
	}
	return out
}

func ParseSegmentPieces(piece io.ReadSeeker) []io.Reader {
	size, _ := piece.Seek(0, io.SeekEnd)
	offset := datasegment.DataSegmentIndexStartOffset(abi.UnpaddedPieceSize(size).Padded())
	piece.Seek(int64(offset), io.SeekStart)
	index, err := datasegment.ParseDataSegmentIndex(piece)
	if err != nil {
		panic(err)
	}
	entries, err := index.ValidEntries()
	if err != nil {
		panic(err)
	}
	out := make([]io.Reader, 0)

	for _, e := range entries {
		buf := bytes.NewBuffer(nil)
		upoff := abi.PaddedPieceSize(e.Offset).Unpadded()
		piece.Seek(int64(upoff), io.SeekStart)
		upSize := abi.PaddedPieceSize(e.Size).Unpadded()
		io.CopyN(buf, piece, int64(upSize))
		out = append(out, buf)
	}
	return out
}
