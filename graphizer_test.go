package graphizer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"reflect"
	"testing"
	"time"
)

type abint struct {
	a int64
	b int64
}

type cd struct {
	c abint
	d int64
}

func Test_parseStruct(t *testing.T) {

	data := []struct {
		in   interface{}
		want []Metric
	}{
		{
			in:   "happy days",
			want: []Metric{Metric{"", "happy days", 0}},
		},
		{
			in:   99,
			want: []Metric{Metric{"", int64(99), 0}},
		},
		{
			in:   struct{ a string }{a: "happy days"},
			want: []Metric{Metric{"a", "happy days", 0}},
		},
		{
			in:   abint{a: 98, b: 99},
			want: []Metric{Metric{"a", int64(98), 0}, Metric{"b", int64(99), 0}},
		},
		{
			in:   cd{c: abint{a: 1, b: 2}, d: 99},
			want: []Metric{Metric{"c.a", int64(1), 0}, Metric{"c.b", int64(2), 0}, Metric{"d", int64(99), 0}},
		},
	}
	for _, v := range data {
		val := ParseStruct(v.in)
		if !reflect.DeepEqual(v.want, val) {
			t.Errorf("want %+v, but got %+v", v.want, val)
			for k, item := range val {
				t.Logf("%+v.  %T %v %T %v", item, item.Value, item.Value, v.want[k].Value, v.want[k].Value)
			}
		}
	}
}

func Test_Connect(t *testing.T) {
	dataChan := make(chan string, 10)
	ln := mockListenAndServe(dataChan, 5555)
	time.Sleep(50 * time.Millisecond)
	// this will panic if there's an error
	g := NewGraphite(TCP, "127.0.0.1:5555")
	g.Close()
	ln.Close()
	time.Sleep(50 * time.Millisecond)

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	t.Log("launching listener")
	// 	ln = mockListenAndServe(dataChan, 5556)
	// }()

	// newGraphite(TCP, "127.0.0.1:5556")

	// g.Close()
	// ln.Close()
}

func Test_Write(t *testing.T) {
	t.Log("testing write")
	ch := make(chan string, 10)
	ln := mockListenAndServe(ch, 5555)
	time.Sleep(50 * time.Millisecond)
	g := NewGraphite("tcp", "localhost:5555")
	now := time.Now().Unix()
	data := []struct {
		in   []Metric
		want []string
	}{
		{
			in:   []Metric{Metric{"path.a.b.c", 1, now}},
			want: []string{fmt.Sprintf("path.a.b.c 1 %d\n", now)},
		},
		{
			in:   []Metric{Metric{"path.a.b.c", 1, now}, Metric{"path.a.b.d", 9, now + 1}},
			want: []string{fmt.Sprintf("path.a.b.c 1 %d\n", now), fmt.Sprintf("path.a.b.d 9 %d\n", now+1)},
		},
	}

	for _, v := range data {
		for _, m := range v.in {
			g.Write(m)
		}
		for i := 0; i < len(v.want); i++ {
			s := <-ch
			if s != v.want[i] {
				t.Errorf("got %s, wanted %s", s, v.want[i])
			}
		}
	}
	g.Close()
	ln.Close()
}

func Test_Send(t *testing.T) {
	t.Log("testing Send")
	ch := make(chan string, 10)
	ln := mockListenAndServe(ch, 5555)
	time.Sleep(50 * time.Millisecond)
	g := NewGraphite("tcp", "localhost:5555")
	now := time.Now().Unix()
	data := []struct {
		in   []Metric
		want []string
	}{
		{
			in:   []Metric{Metric{"path.a.b.c", 1, now}},
			want: []string{fmt.Sprintf("path.a.b.c 1 %d\n", now)},
		},
		{
			in:   []Metric{Metric{"path.a.b.c", 1, now}, Metric{"path.a.b.d", 9, now + 1}},
			want: []string{fmt.Sprintf("path.a.b.c 1 %d\n", now), fmt.Sprintf("path.a.b.d 9 %d\n", now+1)},
		},
	}

	for _, v := range data {
		for _, m := range v.in {
			t.Logf("Sending %s", m.String())
			g.Send(m)
		}
		for i := 0; i < len(v.want); i++ {
			s := <-ch
			t.Logf("Got %s", s)
			if s != v.want[i] {
				t.Errorf("got %s, wanted %s", s, v.want[i])
			}
		}
	}
	g.Close()
	ln.Close()
}

func Test_UDP_Write(t *testing.T) {
	t.Log("testing udp write")
	ch := make(chan string, 10)
	mockUDPListenAndServe(ch, 5555)
	time.Sleep(50 * time.Millisecond)

	g := NewGraphite("udp", ":5555")
	now := time.Now().Unix()
	data := []struct {
		in   []Metric
		want []string
	}{
		{
			in:   []Metric{Metric{"path.a.b.c", 1, now}},
			want: []string{fmt.Sprintf("path.a.b.c 1 %d\n", now)},
		},
		{
			in:   []Metric{Metric{"path.a.b.c", 1, now}, Metric{"path.a.b.d", 9, now + 1}},
			want: []string{fmt.Sprintf("path.a.b.c 1 %d\n", now), fmt.Sprintf("path.a.b.d 9 %d\n", now+1)},
		},
	}

	for _, v := range data {
		for _, m := range v.in {
			g.Write(m)
		}
		for i := 0; i < len(v.want); i++ {
			s := <-ch
			if s != v.want[i] {
				t.Errorf("got: \n\"%s\", wanted \n\"%s\"", s, v.want[i])

			}
		}
	}
	g.Close()
	// ln.Close()
}

func mockListenAndServe(ch chan string, port int) net.Listener {
	var (
		ln  net.Listener
		err error
	)

	ln, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}

	go func(ln net.Listener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(ch chan string, conn net.Conn) {
				reader := bufio.NewReader(conn)
				for {
					conn.SetReadDeadline(time.Now().Add(1 * time.Second))

					line, err := reader.ReadString('\n')
					if err == io.EOF {
						return
					}
					e, ok := err.(net.Error)
					if ok && e.Timeout() {
						continue
					}
					ch <- line
				}
			}(ch, conn)
		}
	}(ln)
	return ln
}

func mockUDPListenAndServe(ch chan string, port int) net.Listener {
	var (
		ln  net.Listener
		err error
	)

	conn, err := net.ListenPacket("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}

	go func(conn net.PacketConn) {
		buff := make([]byte, 512)
		for {
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			if conn == nil {
				return
			}
			_, _, err := conn.ReadFrom(buff)
			if err == io.EOF {
				return
			}
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}

			if bytes.IndexByte(buff, '\n') > 0 {
				buff = buff[0 : bytes.IndexByte(buff, '\n')+1]
			}
			ch <- string(buff)
		}
	}(conn)
	return ln
}
