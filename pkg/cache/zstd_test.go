package cache

import (
	"os"
	"path"
	"testing"

	"github.com/DataDog/zstd"
	"github.com/stretchr/testify/require"
)

// compress with datadog zstd
func zstdDatadogCompress(src []byte) ([]byte, error) { return zstd.Compress(nil, src) }

// decompress with datadog zstd
func zstdDatadogDecompress(src []byte) ([]byte, error) { return zstd.Decompress(nil, src) }

// see https://github.com/klauspost/compress/tree/master/zstd#performance
func BenchmarkZstdCompress(b *testing.B) {
	type args struct {
		in string
	}
	tests := [...]struct {
		name string
		args args
	}{
		{`silesia.tar`, args{"silesia.tar"}},
		{`github-june-2days-2019.json`, args{"github-june-2days-2019.json"}},
		{`gob-stream`, args{"gob-stream"}},
		{`textdata.html`, args{"textdata.html"}},
	}

	for _, tt := range tests {
		b.StopTimer()
		data, err := os.ReadFile(path.Join("fixtures", tt.args.in))
		require.NoError(b, err)
		b.StartTimer()

		b.Run(tt.name+"-zstd", func(b *testing.B) {
			out, err := zstdDatadogCompress(data)
			require.NoError(b, err)
			_ = out
		})

		b.Run(tt.name+"-zskp", func(b *testing.B) {
			out, err := zstdCompress(data)
			require.NoError(b, err)
			_ = out
		})
	}
}

func BenchmarkZstdDecompress(b *testing.B) {
	type args struct {
		in string
	}
	tests := [...]struct {
		name string
		args args
	}{
		{`silesia.tar`, args{"silesia.tar.zst"}},
		{`github-june-2days-2019.json`, args{"github-june-2days-2019.json.zst"}},
		{`gob-stream`, args{"gob-stream.zst"}},
		{`textdata.html`, args{"textdata.html.zst"}},
	}

	for _, tt := range tests {
		b.StopTimer()
		data, err := os.ReadFile(path.Join("fixtures", tt.args.in))
		require.NoError(b, err)
		b.StartTimer()

		b.Run(tt.name+"-zstd", func(b *testing.B) {
			out, err := zstdDatadogDecompress(data)
			require.NoError(b, err)
			_ = out
		})

		b.Run(tt.name+"-zskp", func(b *testing.B) {
			out, err := zstdDecompress(data)
			require.NoError(b, err)
			_ = out
		})
	}
}
