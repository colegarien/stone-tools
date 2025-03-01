package main

import (
	"fmt"
	"os"
	"stone-tools/view"
)

func main() {
	err := view.Run()
	if err != nil {
		fmt.Printf("An Error Occurred: %v", err)
		os.Exit(1)
	}
}
