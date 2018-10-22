package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type responseHeaderStruct struct {
	Organisation []interface{}
	Page         int `json:"page"`
	Size         int `json:"size"`
	TotalPages   int `json:"totalPages"`
	TotalSize    int `json:"totalSize"`
}

type payloadType struct {
	SourceID    string `json:"SourceID"`
	MaxPages    int    `json:"MaxPages"`
	PageSize    int    `json:"PageSize"`
	AsyncSaving bool   `json:"AsyncSaving"`
}

var payload payloadType
var wg sync.WaitGroup
var dummyPayloadFileName = "dummyPayload.json"
var bootTime int64
var invocations int

// main
func main() {
	printMemUsage("main entrypoint 🍾")
	bootTime = time.Now().Unix()
	defer printFinalStatistics()

	if !loadEnvVariables() {
		fmt.Printf("Unable to read .ENV file 💥 \n")
		return
	}

	checkEnvironment()
	greetings()

	fmt.Printf("We may think I'm migrating this from old languages to improve stability and performance\n")
	fmt.Printf("The truth is I just wanted support for emojis.... 🚀 🤠 🍍\n")
	if isAWS {
		// AWS lambda will add the payload to the handler call, we just need to specify the handler fn name...
		lambda.Start(Handler)
	} else {
		// because we are calling the handler manually, we add the payload (a dummy one)
		err := loadDummyPayload()
		if err != nil {
			fmt.Printf("Unable to load dummy payload: %s 💥 💥 💥 \n", err)
			return
		}
		Handler(payload)
	}
	printMemUsage("main exit 🍾 🕺 🍾 ")

}

// Handler doc block....
func Handler(payload payloadType) {
	invocations++
	printMemUsage("Handler entrypoint")

	if payload.MaxPages > 0 {
		fmt.Printf("Max pages to process %v\n", payload.MaxPages)
	}

	fmt.Printf("Async saving 🔀  flag is %v\n", strings.ToUpper(strconv.FormatBool(payload.AsyncSaving)))

	heartbeatKeepAlive(payload.SourceID)
	defer heartbeatEnd()

	switch payload.SourceID {
	case "GRTUKRI":
		importGRTUKRI()
		break
	case "PUBMED":
		importPUBMED()
		break
	default:
		// NOT FOUND!
		fmt.Printf("Source ID %s not found 💥💥 \n", payload.SourceID)
		return
	}

}
