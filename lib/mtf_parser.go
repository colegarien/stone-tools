package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type MtfVirtualFile struct {
	Offset    uint32
	TotalSize uint32
	FileName  string
}

func ExtractVirtualFile(mtfFile io.ReadSeeker, virtualFile MtfVirtualFile) []byte {
	mtfFile.Seek(int64(virtualFile.Offset), io.SeekStart)

	var compressionTag uint32
	err := binary.Read(mtfFile, binary.LittleEndian, &compressionTag)
	if err != nil {
		panic(err)
	}

	if compressionTag != 0xbadbeaf && compressionTag != 0xbadbeae && compressionTag != 0xbadbeaa {
		fmt.Printf("skipping decompress for tag %x", compressionTag)
		// just read data uncompressed
		mtfFile.Seek(int64(virtualFile.Offset), io.SeekStart)

		fileContent := make([]byte, virtualFile.TotalSize)
		mtfFile.Read(fileContent)

		return fileContent
	}

	var compressedSize uint32
	binary.Read(mtfFile, binary.LittleEndian, &compressedSize)

	// grab all the compressed data including the header
	mtfFile.Seek(int64(virtualFile.Offset+compressedSize), io.SeekStart)
	// compressedData := make([]byte, compressedSize)
	// mtfFile.Read(compressedData)

	var currentCrc uint32
	binary.Read(mtfFile, binary.LittleEndian, &currentCrc)

	// decompress logic
	if compressedSize <= 8 {
		// no data available
		return []byte{}
	}

	var sizeOfHeader uint32 = 8 // compressed tag + compressed size fields
	if compressedSize != 0 {
		// paranoid check for 12 byte header available before proceeding to funk; current suspcious is if something is zero bytes, this helps pseudo leaves room for the CRC or something?
		sizeOfHeader = sizeOfHeader + 1 // 9 byte header
		if compressedSize-sizeOfHeader != 0 {
			sizeOfHeader = sizeOfHeader + 1 // 10 byte header

			if compressedSize-sizeOfHeader != 0 {
				sizeOfHeader = sizeOfHeader + 1 // 11 byte header
				if compressedSize-sizeOfHeader != 0 {
					sizeOfHeader = sizeOfHeader + 1 // 12 byte header
				}
			}
		}
	}

	mtfFile.Seek(int64(virtualFile.Offset+sizeOfHeader), io.SeekStart)
	decompressedFile := Decompress(mtfFile, compressedSize-sizeOfHeader)

	newCrc := CRC32(bytes.NewReader(decompressedFile), uint64(virtualFile.TotalSize))
	if currentCrc != newCrc {
		fmt.Printf("mismatching crc %d vs %d\r\n", currentCrc, newCrc)
	} else {
		fmt.Printf("same crcs %d and %d\r\n", currentCrc, newCrc)
	}

	return decompressedFile
}
