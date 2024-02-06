package main

import (
    "MyOwnRedis/internal/redisObject"
    "fmt"
    "log"
    "net"
)

const RedisDefaultPort = 6378

func main() {
    // Create a tcp server that listens to port 6379 through TCP.
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", RedisDefaultPort))
    if err != nil {
        panic(err)
    }
    fmt.Println("My Redis server listening on port 6379...")

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Println(err)
            continue
        }

        // This allows multiple client connections.
        go func() {
            defer func(conn net.Conn) {
                _ = conn.Close()
            }(conn)

            for {
                buf := make([]byte, 1024)
                n, err := conn.Read(buf)
                if err != nil {
                    // Send error message back to the client.
                    // We don't close the connection.
                    break
                }

                // Print the incoming data
                fmt.Printf("Received: %s\n", buf)
                _, _ = redisObject.Deserialize(buf)

                // Close the connection on receiving specific commands.
                if string(buf[:n]) == "quit" {
                    return
                } else {
                    ok := []byte("+PONG\r\n")
                    _, err = conn.Write([]byte(ok))
                    if err != nil {

                    }
                }
            }
        }()
    }
    return
}
