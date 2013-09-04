Graphizer
=========

Graphizer is a go library that talks to graphite.

Usage
=====
```
g := graphizer.newGraphite("udp", "localhost:5555")
g.Send(graphite.Metric{"path.to.metric", 123, time.Now().Unix()})
```

Metrics have the following structure:
```
type Metric struct {
  Path      string
  Value     interface{}
  Timestamp int64
}
```

There are three methods to push data onto graphite, `Write` israw access to the socket, whereas `Send` pushes data onto a channel, and then delivers the metrics async.  `SendStruct`  will serialize a struct into an array of graphite metrics and send them out over the wire.
see: 
- `func (g *Graphite) Write(m Metric) error`
- `func (g *Graphite) Send(m Metric)`
- 

