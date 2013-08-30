package graphizer

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"time"
)

func main() {
	fmt.Println("pork pies")
}

const (
	TCP = "tcp"
	UDP = "udp"
)

//
type Graphite struct {
	proto          string
	endpoint       string
	conn           net.Conn
	ch             chan Metric
	connectTimeout time.Duration
}

// create a new Graphite struct and connect to the underlying connection.
// If the connection isn't immediately successful, we timeout, we sleep, and we retry.
// if we are underbale to connect after 5 attempts, we panic.
func newGraphite(proto, endpoint string) *Graphite {
	g := &Graphite{
		proto:          proto,
		endpoint:       endpoint,
		ch:             make(chan Metric),
		connectTimeout: 30 * time.Second}
	g.getConnection()
	go g.sender()
	return g
}

func (g *Graphite) getConnection() {
	var (
		err   error
		wait  = 20 * time.Second
		count = 0
	)
	log.Printf("Connecting to %s", g.endpoint)
	g.conn, err = net.DialTimeout(g.proto, g.endpoint, g.connectTimeout)
	for err != nil {
		count++
		if count > 5 {
			panic(fmt.Sprintf("Cannot connect to %s", g.endpoint))
		}
		if wait < 60*time.Second {
			wait += 5 * time.Second
		}
		log.Printf("Connection timed out, trying again to connect to %s in %f seconds", g.endpoint, wait.Seconds())
		time.Sleep(wait)
		g.conn, err = net.DialTimeout("tcp", g.endpoint, g.connectTimeout)
	}
}

// close the underlying connection to graphite
func (g *Graphite) Close() {
	g.conn.Close()
}

func (g *Graphite) sender() {
	for m := range g.ch {
		g.Write(m)
	}
}

// Send a single metric to graphite
func (g *Graphite) Send(m Metric) {
	g.ch <- m
}

// Transform a struct into an array of metrics and send them to graphite/
func (g *Graphite) SendStruct(obj interface{}) {
	for _, v := range ParseStruct(obj) {
		g.Send(v)
	}
}

// Write a metric to the underlying conn, and return an error
func (g *Graphite) Write(m Metric) error {
	_, err := g.conn.Write([]byte(m.String()))
	return err
}

// turn a struct into an array of metrics
func ParseStruct(x interface{}) []Metric {
	v := reflect.ValueOf(x)
	return parseStruct(v, make([]Metric, 0), "")
}

func parseStruct(v reflect.Value, acc []Metric, pathSoFar string) []Metric {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			var path string
			if pathSoFar != "" {
				path = pathSoFar + "."
			}
			acc = parseStruct(v.Field(i), acc, path+v.Type().Field(i).Name)
		}
	case reflect.Array:
		vals := make([]interface{}, 0)

		for i := 0; i < v.Len(); i++ {
			vals = append(vals, parseValue(v.Field(i)))
		}
		acc = append(acc, Metric{pathSoFar, vals, 0})
	default:
		acc = append(acc, Metric{pathSoFar, parseValue(v), 0})
	}
	return acc
}

func parseValue(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	}
	return nil
}

// the data that is being passed to graphite.
// Path is the graphite path,
// Value is the metric value.  This is an interface, and can be int, float, etc
// Timestamp should normally be set to time.Now().Unix()
type Metric struct {
	Path      string
	Value     interface{}
	Timestamp int64
}

// Serialize a graphite metric
func (m *Metric) String() string {
	if m.Timestamp == 0 {
		m.Timestamp = time.Now().Unix()
	}
	return fmt.Sprintf("%s %v %d\n", m.Path, m.Value, m.Timestamp)
}
