package main

import (
	"fmt"
	"testing"
)

func TestGetRelationships(t *testing.T) {
	files, err := getFiles("result")
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	for _, file := range files {
		fileRelationships := getRelationships(file)

		// Print file relationships
		fmt.Println("\nFile:", file)
		for _, fileRelationship := range fileRelationships {
			fmt.Println("\t", fileRelationship[0], "->", fileRelationship[1])
		}

		// Print number of relationships
		fmt.Println("\tRelationships: ", len(fileRelationships))
	}
}
