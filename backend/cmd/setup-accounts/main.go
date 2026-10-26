// setup-accounts provisions dashboard accounts, never Minecraft accounts.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"minecraft-dashboard/backend/internal/httpapi"
	"os"
	"path/filepath"
)

func main() {
	directory := flag.String("directory", ".dashboard", "Private output directory (must not already contain account files)")
	flag.Parse()
	if err := os.MkdirAll(*directory, 0700); err != nil {
		panic(err)
	}
	records := []map[string]string{}
	credentials := "Minecraft Dashboard initial credentials — store securely, never commit\n"
	for _, role := range []string{"admin", "viewer"} {
		bytes := make([]byte, 24)
		if _, err := rand.Read(bytes); err != nil {
			panic(err)
		}
		password := base64.RawURLEncoding.EncodeToString(bytes)
		hash, err := httpapi.HashPassword(password)
		if err != nil {
			panic(err)
		}
		records = append(records, map[string]string{"username": role, "role": role, "passwordHash": hash})
		credentials += role + ": " + password + "\n"
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		panic(err)
	}
	// Exclusive creation prevents accidental replacement of configured accounts.
	for _, entry := range []struct {
		name string
		data []byte
	}{{"accounts.json", data}, {"initial-credentials.txt", []byte(credentials)}} {
		f, err := os.OpenFile(filepath.Join(*directory, entry.name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			panic(err)
		}
		if _, err = f.Write(entry.data); err != nil {
			f.Close()
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
	}
	fmt.Println("Created private accounts.json and initial-credentials.txt in", *directory)
}
