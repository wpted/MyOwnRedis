package main

import (
    "MyOwnRedis/internal/database/inMemoryDatabase"
    "MyOwnRedis/internal/server"
)

const RedisDefaultPort string = "6379"

func main() {
    db := inMemoryDatabase.New()
    s := server.New("localhost", RedisDefaultPort, db)
    defer s.Close()

    s.AcceptRequest()
    return
}
