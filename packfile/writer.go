package packfile

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"hash"
	"io"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// WritePackFile to write with the given objects
func WritePackFile(objs map[plumbing.Hash]plumbing.EncodedObject, writer io.Writer) ([]byte, error) {

	cw := newSHA160checksumWriter(writer)
	w := bufio.NewWriterSize(cw, 65519)

	err := writePackfileHeader(uint32(len(objs)), w)
	if err != nil {
		return nil, err
	}
	w.Flush()

	// write all objects
	for _, o := range objs {
		if err = writePackfileEntry(w, o); err != nil {
			return nil, err
		}
		w.Flush()
	}
	//w.Flush()

	checksum := cw.Sum()
	err = binary.Write(w, binary.BigEndian, checksum)

	w.Flush()

	return checksum, err
}

func writePackfileHeader(objCount uint32, w io.Writer) (err error) {
	// signature
	if err = binary.Write(w, binary.BigEndian, []byte("PACK")); err == nil {
		// packfile version
		if err = binary.Write(w, binary.BigEndian, uint32(2)); err == nil {
			// object count
			err = binary.Write(w, binary.BigEndian, objCount)
		}
	}
	return err
}

func writePackfileEntry(w io.Writer, o plumbing.EncodedObject) (err error) {
	var t byte
	// Write type
	t |= 0x80
	switch o.Type() {
	case plumbing.CommitObject:
		t |= byte(plumbing.CommitObject) << 4
	case plumbing.TreeObject:
		t |= byte(plumbing.TreeObject) << 4
	case plumbing.BlobObject:
		t |= byte(plumbing.BlobObject) << 4
	case plumbing.TagObject:
		t |= byte(plumbing.TagObject) << 4
	}
	// Write size
	t |= byte(uint64(o.Size()) &^ 0xfffffffffffffff0)
	sz := o.Size() >> 4
	szb := make([]byte, 16)
	n := binary.PutUvarint(szb, uint64(sz))
	szb = szb[0:n]
	w.Write(append([]byte{t}, szb...))

	// Compress data and write
	zw := zlib.NewWriter(w)
	defer zw.Close()

	or, _ := o.Reader()
	defer or.Close()

	if _, err = io.Copy(zw, or); err == nil {
		zw.Flush()
	}

	return err
}

type checksumWriter struct {
	hash   hash.Hash
	writer io.Writer
}

func newSHA160checksumWriter(w io.Writer) *checksumWriter {
	return &checksumWriter{hash: sha1.New(), writer: w}
}

func (w *checksumWriter) Write(p []byte) (n int, err error) {
	w.hash.Write(p)
	return w.writer.Write(p)
}

func (w *checksumWriter) Sum() []byte {
	return w.hash.Sum(nil)
}
