# kovey network by golang
### Description
#### This is a network library with golang
### Usage
    go get -u github.com/kovey/network-go
### Example Server
```golang
    package main

    import (
        "github.com/kovey/network-go/connection"
        "github.com/kovey/network-go/example"
        "github.com/kovey/network-go/server"
        "encoding/json"
    )

    func main() {
        config := server.Config{}
        packet := connection.PacketConfig{}
        packet.HeaderLength = 4
        packet.BodyLenOffset = 0
        packet.BodyLenLen = 4
        config.PConfig = packet
        config.Host = "127.0.0.1"
        config.Port = 9911

        serv := server.NewServer(config)
        serv.SetService(server.NewTcpService(1024))
        serv.SetHandler(&Handler{})
        serv.Run()
    }

    type Handler struct {
    }

    func (h *Handler) Connect(conn connection.IConnection) error {
        fmt.Printf("new connection[%d]\n", conn.FD())
        return nil
    }

    func (h *Handler) Receive(context *server.Context) error {
        fmt.Printf("%+v\n", context.Pack())
        fmt.Printf("connection[%d]", context.Connection().FD())
        context.Connection().Write(context.Pack())
        return nil
    }

    func (h *Handler) Close(conn connection.IConnection) error {
        fmt.Printf("connection[%d] close \n", conn.FD())
        return nil
    }

    func (h *Handler) Packet(buf []byte) (connection.IPacket, error) {
        p := &Packet{}
        err := p.Unserialize(buf[4:])
        if err != nil {
            return nil, err
        }

        return p, nil
    }

    type Packet struct {
        Action int    `json:"action"`
        Name   string `json:"name"`
        Age    int    `json:"age"`
    }

    func (p *Packet) Serialize() []byte {
        info, err := json.Marshal(p)
        if err != nil {
            return nil
        }

        return append(connection.Int32ToBytes(int32(len(info))), info...)
    }

    func (p *Packet) Unserialize(buf []byte) error {
        return json.Unmarshal(buf, p)
    }

```

### Example Client

```golang
    package main

    import (
        "github.com/kovey/network-go/client"
        "github.com/kovey/network-go/connection"
        "github.com/kovey/network-go/example"
        "encoding/json"
    )

    func main() {
        cli := client.NewClient(connection.PacketConfig{HeaderLength: 4, BodyLenOffset: 0, BodyLenLen: 4})
        cli.SetService(client.NewTcp())
        cli.SetHandler(&Handler{})
        err := cli.Dial("127.0.0.1", 9911)
        if err != nil {
            panic(err)
        }

        pack := &Packet{}
        pack.Action = 1000
        pack.Name = "kovey"
        pack.Age = 18

        cli.Send(pack)

        cli.Loop()
    }

    type Handler struct {
    }

    func (h *Handler) Packet(buf []byte) (connection.IPacket, error) {
        p := &Packet{}
        err := p.Unserialize(buf[4:])
        if err != nil {
            return nil, err
        }

        return p, nil
    }

    func (h *Handler) Receive(pack connection.IPacket, cli *client.Client) error {
        fmt.Printf("%+v\n", pack)
        time.AfterFunc(10*time.Second, func() {
            cli.Send(pack)
        })

        return nil
    }

    func (h *Handler) Idle(cli *client.Client) error {
        return nil
    }

    func (h *Handler) Try(cli *client.Client) bool {
        return false
    }

    type Packet struct {
        Action int    `json:"action"`
        Name   string `json:"name"`
        Age    int    `json:"age"`
    }

    func (p *Packet) Serialize() []byte {
        info, err := json.Marshal(p)
        if err != nil {
            return nil
        }

        return append(connection.Int32ToBytes(int32(len(info))), info...)
    }

    func (p *Packet) Unserialize(buf []byte) error {
        return json.Unmarshal(buf, p)
    }
```
