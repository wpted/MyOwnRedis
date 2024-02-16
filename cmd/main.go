package main

import (
    "MyOwnRedis/internal/database/inMemoryDatabase"
    "MyOwnRedis/internal/server"
    "fmt"
)

const RedisDefaultPort = 6379

func main() {
    db := inMemoryDatabase.New()
    srv := server.New(fmt.Sprintf("localhost:%d", RedisDefaultPort), db)
    err := srv.Run()
    if err != nil {
        panic(err)
    }

    err = srv.Close()

    if err != nil {
        panic(err)
    }
}
