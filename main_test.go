package main_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/ipfs/go-unixfsnode/data/builder"
	"github.com/ipld/go-car/v2"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	parse "github.com/ipld/go-ipld-prime/traversal/selector/parse"
	mkpiece "github.com/willscott/mkpiece/lib"
)

func TestMain(t *testing.T) {
	td := t.TempDir()
	buf1 := bytes.NewBuffer(nil)
	io.CopyN(buf1, rand.Reader, 1024*1024*3)
	if err := mkCar(td+"/carOne.car", buf1); err != nil {
		t.Fatal(err)
	}

	buf2 := bytes.NewBuffer(nil)
	io.CopyN(buf2, rand.Reader, 1024*1024*3)
	if err := mkCar(td+"/carTwo.car", buf2); err != nil {
		t.Fatal(err)
	}

	// combine
	parts := make([]io.ReadSeeker, 0)
	f1, _ := os.Open(td + "/carOne.car")
	parts = append(parts, f1)
	f2, _ := os.Open(td + "/carTwo.car")
	parts = append(parts, f2)
	piece := mkpiece.MakeDataSegmentPiece(parts)
	totalBytes, _ := io.ReadAll(piece)
	f1.Close()
	f2.Close()
	fmt.Printf("total bytes: %d\n", len(totalBytes))
	os.WriteFile(td+"/combined.dat", totalBytes, 0666)

	// now parse.
	fc, _ := os.Open(td + "/combined.dat")
	sgmets := mkpiece.ParseSegmentPieces(fc)
	for i, s := range sgmets {
		dat, _ := io.ReadAll(s)
		os.WriteFile(td+fmt.Sprintf("/recovered_%d.dat", i), dat, 0666)
	}

	if !areSame(td+"/carOne.car", td+"/recovered_0.dat") {
		t.Fatal("carOne.car != recovered_0.dat")
	}
	if !areSame(td+"/carTwo.car", td+"/recovered_1.dat") {
		t.Fatal("carTwo.car != recovered_1.dat")
	}
}

func mkCar(name string, data *bytes.Buffer) error {
	tempStore := memstore.Store{Bag: make(map[string][]byte)}
	ls := cidlink.DefaultLinkSystem()
	ls.SetWriteStorage(&tempStore)
	ls.SetReadStorage(&tempStore)
	rt, _, err := builder.BuildUnixFSFile(data, "default", &ls)
	if err != nil {
		return err
	}
	rtCid := rt.(cidlink.Link).Cid

	carW, err := car.NewSelectiveWriter(context.Background(), &ls, rtCid, parse.CommonSelector_ExploreAllRecursively)
	if err != nil {
		return err
	}

	carOne, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	carW.WriteTo(carOne)
	carOne.Close()
	return nil
}

func areSame(a, b string) bool {
	aBytes, _ := os.ReadFile(a)
	bBytes, _ := os.ReadFile(b)
	// b should be prefixed by a
	if !bytes.Equal(aBytes, bBytes[0:len(aBytes)]) {
		return false
	}
	zeros := make([]byte, len(bBytes)-len(aBytes))
	if !bytes.Equal(bBytes[len(aBytes):], zeros) {
		return false
	}
	return true
}
