package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	ex, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	}

	exPath := filepath.Dir(ex)
	fmt.Println(exPath)
}
