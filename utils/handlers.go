package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

func HandleShutdown(stop <-chan os.Signal) {
	select {
	case <-stop:
		fmt.Println("Shutting down...")
	case <-time.After(30 * time.Second):
		log.Fatal("Shutdown timed out")
	}
}

func HandleErr(err error) {
	fmt.Println(err)
}
