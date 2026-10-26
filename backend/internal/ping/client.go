// Package ping implements the read-only Java server-list status protocol.
package ping

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type Status struct {
	Version     string    `json:"version"`
	Protocol    int       `json:"protocol"`
	Description string    `json:"description"`
	Online      int       `json:"online"`
	Max         int       `json:"max"`
	Sample      []string  `json:"sample"`
	LatencyMS   int64     `json:"latencyMs"`
	CollectedAt time.Time `json:"collectedAt"`
}

func Query(ctx context.Context, address string) (Status, error) {
	var result Status
	host, portText, e := net.SplitHostPort(address)
	if e != nil {
		return result, e
	}
	port, e := strconv.ParseUint(portText, 10, 16)
	if e != nil || len(host) > 255 {
		return result, fmt.Errorf("invalid status address")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	c, e := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if e != nil {
		return result, e
	}
	defer c.Close()
	deadline, _ := ctx.Deadline()
	c.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { c.Close() })
	defer stop()
	handshake := append([]byte{0}, varint(767)...)
	handshake = append(handshake, varint(len(host))...)
	handshake = append(handshake, host...)
	handshake = append(handshake, byte(port>>8), byte(port), 1)
	if e = writePacket(c, handshake); e != nil {
		return result, e
	}
	if e = writePacket(c, []byte{0}); e != nil {
		return result, e
	}
	packet, e := readPacket(c)
	if e != nil {
		return result, e
	}
	r := bytes.NewReader(packet)
	id, e := readVarint(r)
	if e != nil || id != 0 {
		return result, fmt.Errorf("invalid status packet")
	}
	size, e := readVarint(r)
	if e != nil || size != r.Len() {
		return result, fmt.Errorf("invalid status JSON size")
	}
	payload := make([]byte, size)
	io.ReadFull(r, payload)
	var response struct {
		Version struct {
			Name     string
			Protocol int
		}
		Players struct {
			Online *int
			Max    *int
			Sample []struct{ Name string }
		}
		Description json.RawMessage
	}
	if e = json.Unmarshal(payload, &response); e != nil {
		return result, e
	}
	if response.Players.Online == nil || response.Players.Max == nil || *response.Players.Online < 0 || *response.Players.Max < 0 || response.Version.Name == "" {
		return result, fmt.Errorf("incomplete status response")
	}
	result.Version = response.Version.Name
	result.Protocol = response.Version.Protocol
	result.Online = *response.Players.Online
	result.Max = *response.Players.Max
	result.Description = description(response.Description, 0)
	result.Sample = []string{}
	for _, p := range response.Players.Sample {
		if len(result.Sample) >= 100 {
			break
		}
		if len(p.Name) <= 64 {
			result.Sample = append(result.Sample, p.Name)
		}
	}
	sent := time.Now()
	ping := make([]byte, 9)
	ping[0] = 1
	binary.BigEndian.PutUint64(ping[1:], uint64(sent.UnixNano()))
	if e = writePacket(c, ping); e != nil {
		return Status{}, e
	}
	pong, e := readPacket(c)
	if e != nil {
		return Status{}, e
	}
	if !bytes.Equal(ping, pong) {
		return Status{}, fmt.Errorf("invalid status pong")
	}
	result.LatencyMS = time.Since(sent).Milliseconds()
	result.CollectedAt = time.Now().UTC()
	return result, nil
}
func varint(n int) []byte {
	b := []byte{}
	for n >= 128 {
		b = append(b, byte(n)|128)
		n >>= 7
	}
	return append(b, byte(n))
}
func readVarint(r io.Reader) (int, error) {
	n := uint32(0)
	for i := 0; i < 5; i++ {
		var b [1]byte
		if _, e := io.ReadFull(r, b[:]); e != nil {
			return 0, e
		}
		if i == 4 && b[0] > 7 {
			return 0, fmt.Errorf("invalid VarInt")
		}
		n |= uint32(b[0]&127) << uint(7*i)
		if b[0]&128 == 0 {
			return int(n), nil
		}
	}
	return 0, fmt.Errorf("invalid VarInt")
}
func readPacket(r io.Reader) ([]byte, error) {
	n, e := readVarint(r)
	if e != nil {
		return nil, e
	}
	if n < 1 || n > 1024*1024 {
		return nil, fmt.Errorf("status frame exceeds limit")
	}
	b := make([]byte, n)
	_, e = io.ReadFull(r, b)
	return b, e
}
func writePacket(w io.Writer, b []byte) error {
	packet := append(varint(len(b)), b...)
	n, e := w.Write(packet)
	if e == nil && n != len(packet) {
		return io.ErrShortWrite
	}
	return e
}
func description(raw json.RawMessage, depth int) string {
	if depth > 8 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var c struct {
		Text  string
		Extra []json.RawMessage
	}
	if json.Unmarshal(raw, &c) == nil {
		s = c.Text
		for _, part := range c.Extra {
			s += description(part, depth+1)
			if len(s) > 4096 {
				return s[:4096]
			}
		}
		return s
	}
	var list []json.RawMessage
	if json.Unmarshal(raw, &list) == nil {
		for _, part := range list {
			s += description(part, depth+1)
			if len(s) > 4096 {
				return s[:4096]
			}
		}
	}
	return s
}
