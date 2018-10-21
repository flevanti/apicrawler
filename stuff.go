package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/joho/godotenv"
)

var isLambda bool

// checkEnvironment
func checkEnvironment() {
	if len(os.Getenv("AWS_REGION")) != 0 {
		isLambda = true
	} else {
		isLambda = false
	}
}

// greetings
func greetings() {
	if isLambda {
		fmt.Printf("Hello Jeff 🎁   \n")
	} else {
		fmt.Printf("Greetings Professor Falken 🚀   \n")
	}
}

// getDT
func getDT() string {
	// time date formatting...
	// https://golang.org/src/time/format.go
	return time.Now().Format("2006-01-02 15:04:05.0000")
}

// loadEnvVariables
func loadEnvVariables() bool {
	//read .env variables
	if err := godotenv.Load(); err != nil {
		return false
	}
	return true
}

// printMemUsage
func printMemUsage(context string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("MiB usage malloc %v  tot malloc %v  Sys %v  GC %v  (%s)\n",
		b2KiB(m.Alloc), b2KiB(m.TotalAlloc), b2KiB(m.Sys), m.NumGC, context)
}

// b2KiB
func b2KiB(b uint64) uint64 {
	return b / 1024
}

// loadDummyPayload
func loadDummyPayload() error {
	var content string
	var err error
	content = loadDummyPayloadFile()
	err = json.Unmarshal([]byte(content), &payload)
	return err
}

// loadDummyPayloadFile .
func loadDummyPayloadFile() string {
	var content []byte
	var err error
	if fileExists(dummyPayloadFileName) {
		content, err = ioutil.ReadFile(dummyPayloadFileName)
		if err != nil {
			fmt.Printf("Unable to find dummy payload %s\n", dummyPayloadFileName)
			return "{}"
		}
	}
	return string(content)
}

// fileExists
func fileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}

	return false
}

func getInvocations() int {
	return invocations
}

func getUnixTimestamp() int64 {
	return time.Now().Unix()
}

func printFinalStatistics() {
	fmt.Printf("This function existed from %v to %v (%v seconds) ⏰  \n",
		bootTime,
		getUnixTimestamp(),
		(getUnixTimestamp() - bootTime))
	fmt.Printf("It has been invoked %v times\n", getInvocations())

}
