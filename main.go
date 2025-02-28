package main

import (
	"fmt"
	"os"
	"path/filepath"
	"stone-tools/lib"
	"stone-tools/view"
)

func main() {
	// extractAllFiles("music.MTF")
	err := view.Run()
	if err != nil {
		fmt.Printf("An Error Occurred: %v", err)
		os.Exit(1)
	}
}

func extractAllFiles(mtfFilePath string) {
	mtfFile, err := os.Open(mtfFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer mtfFile.Close()

	archive, err := lib.ScanMtfFile(mtfFile)
	if err != nil {
		fmt.Println("Error scanning mtf file:", err)
		return
	}

	for _, virtualFile := range archive.VirtualFiles {
		extractedFile, err := lib.ExtractVirtualFile(mtfFile, virtualFile)
		if err != nil {
			fmt.Printf("Error extracting file `%s`: %+v\r\n", virtualFile.FileName, err)
			continue
		}

		writePath := filepath.Join("out", virtualFile.FileName)
		fmt.Printf("Writing `%s` (%d bytes)...\r\n", writePath, len(extractedFile))

		os.MkdirAll(filepath.Dir(writePath), os.ModePerm)
		err = os.WriteFile(writePath, extractedFile, os.ModePerm)
		if err != nil {
			fmt.Printf("Error writing extracted file `%s`: %+v\r\n", virtualFile.FileName, err)
			continue
		}
	}
}
