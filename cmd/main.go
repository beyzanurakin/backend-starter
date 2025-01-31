package main

import (
    "context"
    "fmt"
    "github.com/beyzanurakin/backend-starter/pkg/logger"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    logger.InitLogger()
    logger.Info("Application is starting...")

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        <-stop
        logger.Info("Received shutdown signal. Cleaning up...")
        cancel()
    }()

    <-ctx.Done()
    fmt.Println("Cleanup complete, exiting program.")
}



