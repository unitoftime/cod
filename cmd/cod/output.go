package main

import (
	"bytes"
	"fmt"
	"go/format"
	"hash/crc32"
	"io/fs"
	"os"
)

func formatFile(buf *bytes.Buffer) []byte {
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes()
	}
	return formatted
}

func outputFile(filename string, buf *bytes.Buffer) {
	formatted := formatFile(buf)

	// Check to see if the file will change
	oldFile, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading last cod file:", err)
	}

	oldSum := crc32.ChecksumIEEE(oldFile)
	newSum := crc32.ChecksumIEEE(formatted)
	if oldSum == newSum {
		fmt.Println("Skipping Write: Files match")
		return
	}

	err = os.WriteFile(filename, formatted, fs.ModePerm)
	if err != nil {
		panic(err)
	}
}
