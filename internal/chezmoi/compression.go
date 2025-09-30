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

// A CompressionFormat is a compression format.
type CompressionFormat string

// Compression formats.
const (
	CompressionFormatNone  CompressionFormat = ""
	CompressionFormatBzip2 CompressionFormat = "bzip2"
	CompressionFormatGzip  CompressionFormat = "gzip"
	CompressionFormatXz    CompressionFormat = "xz"
	CompressionFormatZstd  CompressionFormat = "zstd"
)

func decompress(compressionFormat CompressionFormat, data []byte) ([]byte, error) {
	switch compressionFormat {
	case CompressionFormatNone:
		return data, nil
	case CompressionFormatBzip2:
		return io.ReadAll(bzip2.NewReader(bytes.NewReader(data)))
	case CompressionFormatGzip:
		gzipReader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(gzipReader)
	case CompressionFormatXz:
		xzReader, err := xz.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(xzReader)
	case CompressionFormatZstd:
		zstdReader, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(zstdReader)
	default:
		return nil, fmt.Errorf("%s: unknown compression format", compressionFormat)
	}
}
