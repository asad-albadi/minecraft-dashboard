package chat

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode"
)

const MaxMessageLength = 256

// TellrawCommand builds the only chat command the dashboard emits. JSON encoding,
// rather than string concatenation, prevents command injection through player text.
func TellrawCommand(message string) (string, error) {
	message = strings.TrimSpace(message)
	if message == "" || len([]rune(message)) > MaxMessageLength { return "", errors.New("message must be 1-256 characters") }
	for _, r := range message { if unicode.IsControl(r) && r != '\n' && r != '\t' { return "", errors.New("message contains a control character") } }
	message = strings.NewReplacer("\n", " ", "\t", " ").Replace(message)
	payload, err := json.Marshal(map[string]string{"text": "[Dashboard Admin] " + message})
	if err != nil { return "", err }
	return "tellraw @a " + string(payload), nil
}
