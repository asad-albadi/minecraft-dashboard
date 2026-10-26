package httpapi

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/argon2"
	"io"
	"os"
	"strings"
)

type account struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	Role         string `json:"role"`
}

const hashPrefix = "$argon2id$v=19$m=65536,t=3,p=2$"

// HashPassword produces the supported Argon2id format with a fresh random salt.
func HashPassword(password string) (string, error) {
	if len(password) < 12 || len(password) > 1024 {
		return "", errors.New("password must contain 12 to 1024 bytes")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, 3, 65536, 2, 32)
	return hashPrefix + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}
func hashParts(hash string) ([]byte, []byte, bool) {
	if !strings.HasPrefix(hash, hashPrefix) {
		return nil, nil, false
	}
	parts := strings.Split(strings.TrimPrefix(hash, hashPrefix), "$")
	if len(parts) != 2 {
		return nil, nil, false
	}
	salt, e1 := base64.RawStdEncoding.DecodeString(parts[0])
	key, e2 := base64.RawStdEncoding.DecodeString(parts[1])
	return salt, key, e1 == nil && e2 == nil && len(salt) == 16 && len(key) == 32
}
func verifyPassword(hash, password string) bool {
	salt, key, ok := hashParts(hash)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare(key, argon2.IDKey([]byte(password), salt, 3, 65536, 2, 32)) == 1
}
func loadAccounts(path string) (map[string]account, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > 1024*1024 {
		return nil, errors.New("accounts file must be private and at most 1 MiB")
	}
	dec := json.NewDecoder(io.LimitReader(f, 1024*1024))
	dec.DisallowUnknownFields()
	var records []account
	if err := dec.Decode(&records); err != nil {
		return nil, err
	}
	var extra any
	if dec.Decode(&extra) != io.EOF {
		return nil, errors.New("trailing accounts data")
	}
	if len(records) == 0 || len(records) > 100 {
		return nil, errors.New("configure 1 to 100 accounts")
	}
	result := make(map[string]account)
	for _, record := range records {
		_, _, valid := hashParts(record.PasswordHash)
		_, duplicate := result[record.Username]
		if record.Username == "" || len(record.Username) > 64 || strings.TrimSpace(record.Username) != record.Username || (record.Role != "admin" && record.Role != "viewer") || !valid || duplicate {
			return nil, errors.New("invalid account configuration")
		}
		result[record.Username] = record
	}
	return result, nil
}
