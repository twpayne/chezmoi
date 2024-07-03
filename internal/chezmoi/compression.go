package chezmoi

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"io"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// A compressionFormat is a compression format.
type compressionFormat string

// Compression formats.
const (
	compressionFormatNone  compressionFormat = ""
	compressionFormatBzip2 compressionFormat = "bzip2"
	compressionFormatGzip  compressionFormat = "gzip"
	compressionFormatXz    compressionFormat = "xz"
	compressionFormatZstd  compressionFormat = "zstd"
)

func decompress(compressionFormat compressionFormat, data []byte) ([]byte, error) {
	switch compressionFormat {
	case compressionFormatNone:
		return data, nil
	case compressionFormatBzip2:
		return io.ReadAll(bzip2.NewReader(bytes.NewReader(data)))
	case compressionFormatGzip:
		gzipReader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(gzipReader)
	case compressionFormatXz:
		xzReader, err := xz.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(xzReader)
	case compressionFormatZstd:
		zstdReader, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(zstdReader)
	default:
		return nil, fmt.Errorf("%s: unknown compression format", compressionFormat)
	}
}
