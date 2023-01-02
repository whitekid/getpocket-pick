
## ZSTD

### compress

```txt
BenchmarkCompress/silesia.tar-zstd-16                                   1  1430222476 ns/op  202547472 B/op        6 allocs/op
BenchmarkCompress/silesia.tar-zskp-16                                   7   145437369 ns/op   31910973 B/op      673 allocs/op
BenchmarkCompress/github-june-2days-2019.json-zstd-16                   1 23044579090 ns/op 6298460336 B/op        5 allocs/op
BenchmarkCompress/github-june-2days-2019.json-zskp-16                   1 14067174654 ns/op 1369520424 B/op   143721 allocs/op
BenchmarkCompress/gob-stream-zstd-16                                    1  7409762715 ns/op 1918869680 B/op        5 allocs/op
BenchmarkCompress/gob-stream-zskp-16                                    1  4717475852 ns/op  692727968 B/op    43849 allocs/op
BenchmarkCompress/textdata.html-zstd-16                        1000000000   0.0004539 ns/op          0 B/op        0 allocs/op
BenchmarkCompress/textdata.html-zskp-16                        1000000000    0.003600 ns/op          0 B/op        0 allocs/op
```

### decompress

```txt
BenchmarkZstdDecompress/silesia.tar-zstd-16                 1000000000      0.1758 ns/op        0 B/op        0 allocs/op
BenchmarkZstdDecompress/silesia.tar-zskp-16                 1000000000      0.2354 ns/op        0 B/op        0 allocs/op
BenchmarkZstdDecompress/github-june-2days-2019.json-zstd-16          1  5000797005 ns/op 6273958088 B/op        6 allocs/op
BenchmarkZstdDecompress/github-june-2days-2019.json-zskp-16          1 12936828784 ns/op 17191837968 B/op      262 allocs/op
BenchmarkZstdDecompress/gob-stream-zstd-16                           1  3098997468 ns/op 13090318840 B/op    58416 allocs/op
BenchmarkZstdDecompress/gob-stream-zskp-16                           1  1371059656 ns/op 4301388680 B/op      138 allocs/op
BenchmarkZstdDecompress/textdata.html-zstd-16               1000000000   0.0000764 ns/op        0 B/op        0 allocs/op
BenchmarkZstdDecompress/textdata.html-zskp-16               1000000000   0.0001815 ns/op        0 B/op        0 allocs/op
```
