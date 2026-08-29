package main

import (
	"fmt"
	"os"

	"github.com/DonaldMurillo/fastr-docs/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "fastr-docs:", err)
		os.Exit(1)
	}
}
