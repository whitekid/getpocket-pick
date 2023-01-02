package cache

import (
	"bytes"
	"io"
	"runtime"

	"github.com/klauspost/compress/zstd"
)

// zstdCompress compress with klauspost zstd
func zstdCompress(src []byte) ([]byte, error) {
	out := bytes.NewBuffer(nil)
	enc, err := zstd.NewWriter(out, zstd.WithEncoderConcurrency(runtime.NumCPU()))
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(enc, bytes.NewReader(src)); err != nil {
		return nil, err
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// zstdDecompress decompress with klauspost zstd
func zstdDecompress(src []byte) ([]byte, error) {
	dec, err := zstd.NewReader(bytes.NewReader(src), zstd.WithDecoderConcurrency(runtime.NumCPU()))
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	out := bytes.NewBuffer(nil)
	if _, err = io.Copy(out, dec); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
