package ping

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestQueryStatusAndPong(t *testing.T) {
	l, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	go func() {
		c, e := l.Accept()
		if e != nil {
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(time.Second))
		for i := 0; i < 2; i++ {
			var size [1]byte
			if _, e = io.ReadFull(c, size[:]); e != nil {
				return
			}
			io.CopyN(io.Discard, c, int64(size[0]))
		}
		s := `{"version":{"name":"1.21","protocol":767},"players":{"online":1,"max":20,"sample":[{"name":"Alex","id":"private"}]},"description":{"text":"Hello","extra":[{"text":" world"}]}}`
		b := append([]byte{0}, testVarint(len(s))...)
		b = append(b, []byte(s)...)
		c.Write(append(testVarint(len(b)), b...))
		var p [10]byte
		io.ReadFull(c, p[:])
		c.Write(p[:])
	}()
	s, e := Query(context.Background(), l.Addr().String())
	if e != nil || s.Version != "1.21" || s.Online != 1 || s.Max != 20 || s.Description != "Hello world" || len(s.Sample) != 1 || s.Sample[0] != "Alex" {
		t.Fatalf("%+v %v", s, e)
	}
}
func testVarint(n int) []byte {
	b := []byte{}
	for n >= 128 {
		b = append(b, byte(n)|128)
		n >>= 7
	}
	return append(b, byte(n))
}
func TestQueryRejectsOversizedFrame(t *testing.T) {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	defer l.Close()
	go func() {
		c, e := l.Accept()
		if e != nil {
			return
		}
		defer c.Close()
		c.Write(testVarint(2 * 1024 * 1024))
	}()
	_, e := Query(context.Background(), l.Addr().String())
	if e == nil {
		t.Fatal("oversize accepted")
	}
}
func TestQueryHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e := Query(ctx, "127.0.0.1:1")
	if e == nil {
		t.Fatal("cancel ignored")
	}
}
