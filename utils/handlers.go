package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func HandleShutdown(stop <-chan os.Signal, delay time.Duration) {
	select {
	case <-stop:
		logrus.Info("Shutting down...")
	case <-time.After(delay):
		logrus.Fatal("Shutdown timed out")
	}
}

func HandleWaitAndShutdown(stop <-chan os.Signal, delay time.Duration) {
	select {
	case <-stop:
		logrus.Info("Shutting down...")
	case <-time.After(delay):
		logrus.Info("Shutdown timed out")
	}
}

func HandleErr(err error) {
	fmt.Println(err)
}
