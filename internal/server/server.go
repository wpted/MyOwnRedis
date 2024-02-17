package server

import (
    "MyOwnRedis/internal/database"
    "MyOwnRedis/internal/database/inMemoryDatabase"
    "MyOwnRedis/internal/redisObject"
    "fmt"
    "net"
)

const TCP = "tcp"

// 1. Create a RedisServer struct
// 2. Create methods Run(), Close() and Evaluate().

type RedisServer struct {
    addr string
    // Passing `net.Listener` by value is idiomatic and aligns with the general practice in Go of passing interface by value.
    l  net.Listener
    db database.MemDb
}

// New creates a new RedisServer.
func New(addr string, db database.MemDb) *RedisServer {
    return &RedisServer{
        addr: addr,
        db:   db,
    }
}

// Run sets up a Redis server that can handle multiple client connections concurrently and respond to client requests asynchronously.
// Each connection is handled in a separate goroutine, allowing the server to remain responsive to new connections and handle them efficiently.
func (r *RedisServer) Run() error {
    var err error
    // Create a socket that can accept incoming connections.
    r.l, err = net.Listen(TCP, r.addr)
    if err != nil {
        return err
    }

    // Forever loop ran here is necessary to continuously accept incoming connections.
    for {
        var conn net.Conn
        // Waiting for incoming connections.
        conn, err = r.l.Accept()
        if err != nil || conn == nil {
            // If anything happens when waiting for connections, process towards the next loop.
            // i.e. Error Accepting or Empty connection ( meaning that there's no incoming connection )
            continue
        }

        // If receive a connection, spawn the connection dealing process with a goroutine.
        go func(conn net.Conn) {
            // Close the connection after we're done dealing with the connection.
            defer func() {
                err := conn.Close()
                if err != nil {
                    fmt.Println(err)
                }
            }()
            // Read data from the connection.
            // For loop here is to enable sequential network reads on the same connection,
            // and will break
            for {
                req := make([]byte, 1024)
                n, err := conn.Read(req)
                if err != nil {

                    break
                }

                // Trim empty bytes.
                req = req[:n]

                // Handle request.
                var response []byte
                response, err = r.handleRequest(req)
                if err != nil {
                    break
                }

                // Write response to the connection (Responding to client).
                if _, err = conn.Write(response); err != nil {
                    break
                }
            }
        }(conn)
    }
}

// Close closes the listener.
// Any blocked `Accept` operations will be unblocked and return errors.
func (r *RedisServer) Close() error {
    return r.l.Close()
}

func (r *RedisServer) evaluate(robj *redisObject.RObj) ([]byte, error) {
    var resp []byte
    switch robj.Command {
    case "ping":
        resp = redisObject.Serialize(redisObject.SimpleStrings, "PONG")
    case "echo":
        resp = redisObject.Serialize(redisObject.SimpleStrings, robj.Content...)
    case "set":
        // Any SET operation will be successful and previous value is discarded.
        // The command should always return '+OK\r\n'.
        r.db.Set(robj.Content[0], robj.Content[1])
        resp = redisObject.Serialize(redisObject.SimpleStrings, "OK")
    case "get":
        value, err := r.db.Get(robj.Content[0])
        if err != nil {
            // The error here can only be clients trying to get from the lrange database.
            resp = redisObject.Serialize(redisObject.SimpleErrors, "ERR WRONGTYPE Operation against a key holding the wrong kind of value")
        } else {
            // Check for nil values.
            if value == inMemoryDatabase.NIL {
                resp = redisObject.Serialize(redisObject.BulkStrings, "-1")
            } else {
                resp = redisObject.Serialize(redisObject.SimpleStrings, value)
            }

            // Do we have to check whether the value is an integer?
        }
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

func (r *RedisServer) handleRequest(request []byte) ([]byte, error) {
    var response []byte

    robj, err := redisObject.Deserialize(request)
    if err != nil {
        response = redisObject.Serialize(redisObject.SimpleErrors, err.Error())
    }
    response, err = r.evaluate(robj)
    if err != nil {

    }

    return response, nil
}
