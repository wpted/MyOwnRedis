package server

import (
    "MyOwnRedis/internal/database"
    "MyOwnRedis/internal/redisObject"
    "fmt"
    "log"
    "net"
)

type RedisServer struct {
    addr string
    l    *net.Listener
    db   database.MemDb
}

func New(host, port string, db database.MemDb) *RedisServer {
    addr := fmt.Sprintf("%s:%s", host, port)
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("Error listening: %#v.\n", err)
    }

    return &RedisServer{
        addr: addr,
        l:    &listener,
        db:   db,
    }
}

func (r *RedisServer) Close() {
    if err := (*r.l).Close(); err != nil {
        log.Fatalf("Error closing connection: %#v.\n", err)
    }
}

func (r *RedisServer) AcceptRequest() {
    for {
        conn, err := (*r.l).Accept()
        if err != nil {
            log.Fatalf("Error accepting: %#v.\n", err)
        }
        // Handle connections with goroutines.
        go func(conn net.Conn) {
            // Create a buffer to hold incoming request.
            reqBuf := make([]byte, 1024)
            // Read the incoming connection into the buffer.
            _, err = conn.Read(reqBuf)
            fmt.Println(string(reqBuf))

            if err != nil {
                log.Println("Error reading request:", err.Error())
            }

            // Do something with the request buffer.
            robj, err := redisObject.Deserialize(reqBuf)
            // Send response back to client.
            var resp []byte
            if err != nil {
                resp = redisObject.Serialize(redisObject.SimpleErrors, err.Error())
            } else {
                resp, err = r.Evaluate(robj)
                if err != nil {
                    resp = redisObject.Serialize(redisObject.SimpleErrors, err.Error())
                } else {
                    if _, err = conn.Write(resp); err != nil {
                        panic(err)
                    }
                }
            }

            // Close connection.
            //if err := conn.Close(); err != nil {
            //    panic(err)
            //}
        }(conn)
    }
}

func (r *RedisServer) Evaluate(robj *redisObject.RObj) ([]byte, error) {
    var resp []byte
    switch robj.Command {
    case "ping":
        resp = redisObject.Serialize(redisObject.SimpleStrings, "PONG")
    case "echo":
        resp = redisObject.Serialize(redisObject.SimpleStrings, robj.Content...)
    case "quit":
    case "set":
    case "get":
    case "del":
    case "exists":
    case "incr":
    case "decr":
    case "save":
    case "load":
    case "lpush":
    case "rpush":
    case "lrange":
    }
    return resp, nil
}
