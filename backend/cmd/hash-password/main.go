// hash-password reads one password from stdin; passwords never appear in arguments.
package main

import (
	"fmt"
	"io"
	"minecraft-dashboard/backend/internal/httpapi"
	"os"
	"strings"
)

func main() {
	data, err := io.ReadAll(io.LimitReader(os.Stdin, 1026))
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read password")
		os.Exit(1)
	}
	hash, err := httpapi.HashPassword(strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(hash)
}
