package server

import (
    "MyOwnRedis/internal/database"
    "MyOwnRedis/internal/database/inMemoryDatabase"
    "MyOwnRedis/internal/redisObject"
    "fmt"
    "net"
    "strconv"
    "sync"
    "time"
)

const TCP = "tcp"

type RedisServer struct {
    addr string
    // Passing `net.Listener` by value is idiomatic and aligns with the general practice in Go of passing interface by value.
    l            net.Listener
    db           database.MemDb
    keysChanged  int
    done         chan struct{}
    saveRoutines map[time.Duration]struct {
        timeCreated time.Time
        done        chan struct{}
    }
    sync.RWMutex
}

// New creates a new RedisServer.
func New(addr string, db database.MemDb) *RedisServer {
    return &RedisServer{
        addr: addr,
        db:   db,
        done: make(chan struct{}),
        saveRoutines: make(map[time.Duration]struct {
            timeCreated time.Time
            done        chan struct{}
        }),
    }
}

// Run sets up a Redis server that can handle multiple client connections concurrently and respond to client requests asynchronously.
// Each connection is handled in a separate goroutine, allowing the server to remain responsive to new connections and handle them efficiently.
func (r *RedisServer) Run() error {
    go r.save(900*time.Second, 1)
    go r.save(300*time.Second, 100)

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
        // Then go on to the next loop.
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
                response = r.handleRequest(req)
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
    r.done <- struct{}{}
    return r.l.Close()
}

func (r *RedisServer) handleRequest(request []byte) []byte {
    var response []byte

    robj, err := redisObject.Deserialize(request)
    if err != nil {
        response = redisObject.Serialize(redisObject.SimpleErrors, "Unknown or disabled command")
    } else {
        switch robj.Command {
        case "ping":
            response = redisObject.Serialize(redisObject.SimpleStrings, "PONG")

        case "echo":
            response = redisObject.Serialize(redisObject.SimpleStrings, robj.Content...)

        case "set":
            // Any SET operation will be successful and previous value is discarded.
            // The command should always return '+OK\r\n'.
            r.db.Set(robj.Content[0], robj.Content[1])

            r.RLock()
            r.keysChanged++
            r.RUnlock()

            response = redisObject.Serialize(redisObject.SimpleStrings, "OK")

            // Expire robj that has time to live.
            if robj.TimeToLive != 0 {
                // Launch a goroutine that waits to expire the object.
                go r.expireRObj(robj)
            }

        case "get":
            value, err := r.db.Get(robj.Content[0])
            if err != nil {
                // The error here can only be clients trying to get from the lrange database.
                response = redisObject.Serialize(redisObject.SimpleErrors, "WRONGTYPE Operation against a key holding the wrong kind of value")
            } else {
                // Check for nil values.
                if value == inMemoryDatabase.Nil {
                    response = redisObject.Serialize(redisObject.BulkStrings, "-1")
                    return response
                }
                response = redisObject.Serialize(redisObject.SimpleStrings, value)
                // Do we have to check whether the value is an integer?
            }

        case "del":
            keysDeleted := r.db.Delete(robj.Content...)
            r.RLock()
            r.keysChanged += keysDeleted
            r.RUnlock()
            // Integer response.
            response = redisObject.Serialize(redisObject.Integers, strconv.Itoa(keysDeleted))

        case "exists":
            // Integer response. 1 for found key, 0 otherwise.
            if r.db.Exists(robj.Content[0]) {
                response = redisObject.Serialize(redisObject.Integers, strconv.Itoa(1))
                return response
            }
            response = redisObject.Serialize(redisObject.Integers, strconv.Itoa(0))

        case "incr":
            value, err := r.db.Increment(robj.Content[0])
            if err != nil {
                response = redisObject.Serialize(redisObject.SimpleErrors, "WRONGTYPE Operation against a key holding the wrong kind of value")
                return response
            }
            r.RLock()
            r.keysChanged++
            r.RUnlock()
            response = redisObject.Serialize(redisObject.Integers, strconv.Itoa(value))

        case "decr":
            value, err := r.db.Decrement(robj.Content[0])
            if err != nil {
                response = redisObject.Serialize(redisObject.SimpleErrors, "WRONGTYPE Operation against a key holding the wrong kind of value")
                return response
            }
            r.RLock()
            r.keysChanged++
            r.RUnlock()
            response = redisObject.Serialize(redisObject.Integers, strconv.Itoa(value))

        case "save":
            if err := r.db.SaveDatabase(); err != nil {
                panic(err)
            }
            if len(robj.Content) != 0 {
                // Setup save options
                go r.save(robj.SaveOptions.CheckCycle, robj.SaveOptions.CheckKeys)
            }
            response = redisObject.Serialize(redisObject.SimpleStrings, "OK")

        case "lpush":
            valuesPushed, err := r.db.LeftPush(robj.Content[0], robj.Content[1:]...)
            if err != nil {
                response = redisObject.Serialize(redisObject.SimpleErrors, "WRONGTYPE Operation against a key holding the wrong kind of value")
                return response
            }
            r.RLock()
            r.keysChanged++
            r.RUnlock()

            response = redisObject.Serialize(redisObject.Integers, strconv.Itoa(valuesPushed))

        case "rpush":
            valuesPushed, err := r.db.RightPush(robj.Content[0], robj.Content[1:]...)
            if err != nil {
                response = redisObject.Serialize(redisObject.SimpleErrors, "WRONGTYPE Operation against a key holding the wrong kind of value")
                return response
            }
            r.RLock()
            r.keysChanged++
            r.RUnlock()
            response = redisObject.Serialize(redisObject.Integers, strconv.Itoa(valuesPushed))

        case "lrange":
            // Need to type check the start and end.
            start, err := strconv.Atoi(robj.Content[1])
            if err != nil {
                response = redisObject.Serialize(redisObject.SimpleErrors, "ERR value is not an integer or out of range")
                return response
            }

            end, err := strconv.Atoi(robj.Content[2])
            if err != nil {
                response = redisObject.Serialize(redisObject.SimpleErrors, "ERR value is not an integer or out of range")
                return response
            }

            var result []string
            result, err = r.db.LRange(robj.Content[0], start, end)
            if err != nil {
                response = redisObject.Serialize(redisObject.SimpleErrors, "WRONGTYPE Operation against a key holding the wrong kind of value")
                return response
            }
            response = redisObject.Serialize(redisObject.Arrays, result...)

        case "command": // This is for starting up, which will never show on the client side.
            response = redisObject.Serialize(redisObject.SimpleStrings, "Hello, Edward's Redis.")
        }
    }
    return response
}

// expireRObj expires a Redis object after reaches time to live.
func (r *RedisServer) expireRObj(robj *redisObject.RObj) {
    // Blocking process.
    select {
    // Choose time.After over time.NewTimer for light-weight purposes.
    case <-time.After(robj.TimeToLive):
        // Will delete when receive from time.After(), unblock process.
        r.db.Delete(robj.Content[0])
    }
}

// save method is responsible for managing periodic saving operations for the Redis server.
// It monitors the number of active save routines and limits them to five.
// If there are already five save routines running, it replaces the earliest one with a new one.
func (r *RedisServer) save(checkCycle time.Duration, checkKeys int) {
    r.Lock()
    if len(r.saveRoutines) >= 5 {
        // Collect keys to delete
        var earliestKey time.Duration
        earliestTime := time.Now()
        for k, t := range r.saveRoutines {
            if t.timeCreated.Before(earliestTime) {
                earliestTime = t.timeCreated
                earliestKey = k
            }
        }
        r.saveRoutines[earliestKey].done <- struct{}{}
        delete(r.saveRoutines, earliestKey)
    }
    now := time.Now()
    saveChan := make(chan struct{})
    r.saveRoutines[checkCycle] = struct {
        timeCreated time.Time
        done        chan struct{}
    }{timeCreated: now, done: saveChan}
    r.Unlock()

    ticker := time.NewTicker(checkCycle)
    initialKey := r.keysChanged

    for {
        select {
        case <-r.done: // Release current goroutine.
            return
        case <-saveChan:
            return
        case <-ticker.C:
            r.Lock()
            if r.keysChanged-initialKey >= checkKeys {
                _ = r.db.SaveDatabase()
                initialKey = r.keysChanged
            }
            r.Unlock()
        }
    }
}
