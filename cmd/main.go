package main

import (
    "MyOwnRedis/internal/database/inMemoryDatabase"
    "MyOwnRedis/internal/server"
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

const RedisDefaultPort = 6379

func main() {
    db := inMemoryDatabase.New()
    srv := server.New(fmt.Sprintf("localhost:%d", RedisDefaultPort), db)

    go func() {
        err := srv.Run()
        if err != nil {
            log.Fatalf("RRedis server start up error: %v", err)
        }
    }()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    // Try receiving from channel sigChan (blocking), main routine blocked.
    <-sigChan
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    // Cancel the context after timeout (sends signal to the Done chan).
    defer cancel()

    if err := srv.Close(shutdownCtx); err != nil {
        log.Fatalf("RRedis server shutdown error: %v", err)
    }
    log.Println("RRedis shutdown complete.")
}
