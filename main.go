package main

import (
	"fmt"
	"os"
	"stone-tools/lib"
)

func main() {
	// file, err := os.Open("music.MTF")
	file, err := os.Open("data.mtf")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	virtualFile := lib.ExtractVirtualFile(file, lib.MtfVirtualFile{
		// Offset:    629,
		// TotalSize: 1440915,
		Offset:    283300,
		TotalSize: 196652,
	})

	os.WriteFile("sandwich.tga", virtualFile, os.ModePerm)
}
