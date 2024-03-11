package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

func HandleShutdown(stop <-chan os.Signal, delay time.Duration) {
	select {
	case <-stop:
		fmt.Println("Shutting down...")
	case <-time.After(delay):
		log.Fatal("Shutdown timed out")
	}
}

func HandleErr(err error) {
	fmt.Println(err)
}
