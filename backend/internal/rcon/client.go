package rcon

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	commandRequest = 2
	commandResponse = 0
	auth = 3
	authResponse = 2
)

type packet struct { ID, Type int32; Body string }
type Client struct { address, password string; timeout time.Duration }

func New(address, password string, timeout time.Duration) *Client { return &Client{address: address, password: password, timeout: timeout} }

func (c *Client) Command(ctx context.Context, command string) (string, error) {
	if c.address == "" || c.password == "" { return "", errors.New("RCON is not configured") }
	dialer := net.Dialer{Timeout: c.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.address)
	if err != nil { return "", fmt.Errorf("connect to RCON: %w", err) }
	defer conn.Close()
	deadline := time.Now().Add(c.timeout); if d, ok := ctx.Deadline(); ok && d.Before(deadline) { deadline = d }; _ = conn.SetDeadline(deadline)
	if err := writePacket(conn, packet{ID: 1, Type: auth, Body: c.password}); err != nil { return "", err }
	authPacket, err := readPacket(conn); if err != nil { return "", err }
	if authPacket.Type != authResponse || authPacket.ID == -1 { return "", errors.New("RCON authentication failed") }
	if err := writePacket(conn, packet{ID: 2, Type: commandRequest, Body: command}); err != nil { return "", err }
	response, err := readPacket(conn); if err != nil { return "", err }
	if response.Type != commandResponse { return "", fmt.Errorf("unexpected RCON response type %d", response.Type) }
	return response.Body, nil
}

func writePacket(w io.Writer, p packet) error { body := append([]byte(p.Body), 0, 0); size := int32(8 + len(body)); raw := make([]byte, 4+size); binary.LittleEndian.PutUint32(raw[0:4], uint32(size)); binary.LittleEndian.PutUint32(raw[4:8], uint32(p.ID)); binary.LittleEndian.PutUint32(raw[8:12], uint32(p.Type)); copy(raw[12:], body); _, err := w.Write(raw); return err }
func readPacket(r io.Reader) (packet, error) { var prefix [4]byte; if _, err := io.ReadFull(r, prefix[:]); err != nil { return packet{}, err }; size := binary.LittleEndian.Uint32(prefix[:]); if size < 10 || size > 65536 { return packet{}, errors.New("invalid RCON packet length") }; raw := make([]byte, size); if _, err := io.ReadFull(r, raw); err != nil { return packet{}, err }; return packet{ID:int32(binary.LittleEndian.Uint32(raw[:4])), Type:int32(binary.LittleEndian.Uint32(raw[4:8])), Body:string(raw[8:len(raw)-2])}, nil }

var listPattern = regexp.MustCompile(`There are (\d+) of a max of (\d+) players online(?::\s*(.*))?`)
func ParseListResponse(response string) (online, max int, names []string, ok bool) { match := listPattern.FindStringSubmatch(response); if match == nil { return 0, 0, nil, false }; online, _ = strconv.Atoi(match[1]); max, _ = strconv.Atoi(match[2]); if len(match) > 3 && strings.TrimSpace(match[3]) != "" { names = strings.Split(match[3], ", ") }; return online, max, names, true }
