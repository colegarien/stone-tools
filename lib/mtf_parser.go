package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type MtfArchive struct {
	VirtualFiles []MtfVirtualFile
}

type MtfVirtualFile struct {
	Offset    uint32
	TotalSize uint32
	FileName  string
}

func ScanMtfFile(mtfFile io.ReadSeeker) (MtfArchive, error) {
	mtfFile.Seek(0, io.SeekStart)

	var numberOfVirtualFiles uint32
	err := binary.Read(mtfFile, binary.LittleEndian, &numberOfVirtualFiles)
	if err != nil {
		return MtfArchive{}, err
	}

	archive := MtfArchive{
		VirtualFiles: make([]MtfVirtualFile, 0, numberOfVirtualFiles),
	}
	for range numberOfVirtualFiles {
		var nameLength uint32
		err = binary.Read(mtfFile, binary.LittleEndian, &nameLength)
		if err != nil {
			return archive, err
		}

		var name = make([]byte, nameLength)
		_, err = mtfFile.Read(name)
		if err != nil {
			return archive, err
		}

		var offset uint32
		err = binary.Read(mtfFile, binary.LittleEndian, &offset)
		if err != nil {
			return archive, err
		}

		var totalSize uint32
		err = binary.Read(mtfFile, binary.LittleEndian, &totalSize)
		if err != nil {
			return archive, err
		}

		archive.VirtualFiles = append(archive.VirtualFiles, MtfVirtualFile{
			Offset:    offset,
			TotalSize: totalSize,
			FileName:  filepath.Clean(strings.Trim(string(name), "\x00")),
		})
	}

	return archive, nil
}

func ExtractVirtualFile(mtfFile io.ReadSeeker, virtualFile MtfVirtualFile) ([]byte, error) {
	_, err := mtfFile.Seek(int64(virtualFile.Offset), io.SeekStart)
	if err != nil {
		return nil, err
	}

	var compressionTag uint32
	err = binary.Read(mtfFile, binary.LittleEndian, &compressionTag)
	if err != nil {
		return nil, err
	}

	if compressionTag != 0xbadbeaf && compressionTag != 0xbadbeae && compressionTag != 0xbadbeaa {
		// just read data uncompressed
		mtfFile.Seek(int64(virtualFile.Offset), io.SeekStart)

		fileContent := make([]byte, virtualFile.TotalSize)
		mtfFile.Read(fileContent)

		return fileContent, nil
	}

	var compressedSize uint32
	err = binary.Read(mtfFile, binary.LittleEndian, &compressedSize)
	if err != nil {
		return nil, err
	}

	// skip over compressed data and grab the current crc
	var currentCrc uint32
	_, err = mtfFile.Seek(int64(virtualFile.Offset+compressedSize), io.SeekStart)
	if err != nil {
		return nil, err
	}

	err = binary.Read(mtfFile, binary.LittleEndian, &currentCrc)
	if err != nil {
		return nil, err
	}

	// decompress logic
	if compressedSize <= 8 {
		// no data available
		return nil, nil
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

	_, err = mtfFile.Seek(int64(virtualFile.Offset+sizeOfHeader), io.SeekStart)
	if err != nil {
		return nil, err
	}

	decompressedFile, err := Decompress(mtfFile, compressedSize-sizeOfHeader)
	if err != nil {
		return nil, err
	}

	newCrc := CRC32(bytes.NewReader(decompressedFile), uint64(virtualFile.TotalSize))
	if currentCrc != newCrc {
		fmt.Printf("mismatching crc %d vs %d\r\n", currentCrc, newCrc)
	}

	return decompressedFile, nil
}
