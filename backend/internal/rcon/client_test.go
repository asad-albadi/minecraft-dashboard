package rcon

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestClientAuthenticatesAndReturnsCommandResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil { t.Fatal(err) }
	defer listener.Close()
	go func() {
		conn, _ := listener.Accept(); defer conn.Close()
		authRequest, _ := readPacket(conn)
		if authRequest.Type != auth || authRequest.Body != "secret" { return }
		_ = writePacket(conn, packet{ID: authRequest.ID, Type: authResponse})
		command, _ := readPacket(conn)
		if command.Type == commandRequest && command.Body == "list" { _ = writePacket(conn, packet{ID: command.ID, Type: commandResponse, Body: "There are 2 of a max of 20 players online: Alex, Sam"}) }
	}()
	client := New(listener.Addr().String(), "secret", time.Second)
	response, err := client.Command(context.Background(), "list")
	if err != nil { t.Fatal(err) }
	if response != "There are 2 of a max of 20 players online: Alex, Sam" { t.Fatalf("unexpected response: %q", response) }
}

func TestParseListResponse(t *testing.T) {
	players, max, names, ok := ParseListResponse("There are 2 of a max of 20 players online: Alex, Sam")
	if !ok || players != 2 || max != 20 || len(names) != 2 || names[0] != "Alex" { t.Fatalf("unexpected parse: %d %d %#v %t", players, max, names, ok) }
}
