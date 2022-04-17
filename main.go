package main

import (
	"fmt"
	"os"
)

func main() {
	// Print the command line arguments
	fmt.Println(os.Args)
	// Print the number of command line arguments
	fmt.Println(len(os.Args))

	// Set configuration
	configure()

	switch command := os.Args[1]; command {
	case "convert":
		if len(os.Args) != 4 {
			fmt.Println("Usage: www-graph convert <source-dir> <dest-dir>")
			os.Exit(1)
		}
		html2md(os.Args[2], os.Args[3])
	case "check":
		if len(os.Args) != 2 {
			fmt.Println("Usage: www-graph check")
			os.Exit(1)
		}
		check()
	case "export":
		if len(os.Args) != 4 {
			fmt.Println("Usage: www-graph export <source-dir> <dest-dir>")
			os.Exit(1)
		}
		export(os.Args[2], os.Args[3])
	}
}
