package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

type payloadType struct {
	SourceID                    string `json:"SourceID,omitempty"`
	MaxPages                    int    `json:"MaxPages,omitempty"`
	PageSize                    int    `json:"PageSize,omitempty"`
	AsyncSaving                 bool   `json:"AsyncSaving,omitempty"`
	EndpointsParallelProcessing bool   `json:"EndpointsParallelProcessing,omitempty"`

	Endpoints []struct {
		Name            string `json:"Name"`
		Uri             string `json:"Uri"`
		Collection      string `json:"Collection"`
		ResponseElement string `json:"ResponseElement"`
	} `json:"Endpoints"`
}

var payload payloadType
var dummyPayloadFileName = "dummyPayload.json"
var bootTime int64
var invocations int

// main
func main() {
	printMemUsage("main entrypoint 🍾")
	bootTime = time.Now().Unix()
	defer printFinalStatistics()

	if !loadEnvVariables() {
		fmt.Printf("Unable to read .env file 💥 \n")
		return
	}

	checkEnvironment()
	greetings()

	fmt.Printf("We may think I'm migrating this from old languages to improve stability and performance\n")
	fmt.Printf("The truth is I just wanted support for emojis.... 🚀 🤠 🍍\n")
	if isLambda {
		// AWS lambda will add the payload to the handler call, we just need to specify the handler fn name...
		// If the lambda is the local docker implementation, we need to pass the payload (a dummy one) as an argument
		lambda.Start(Handler)
	} else {
		// because we are calling the handler manually, we load a payload (a dummy one)
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
func Handler(payloadLocalScope payloadType) {
	// if the handler was called by lambda we have the payload passed
	// as a parameter but not in the global var... time to take care of it...
	if isLambda {
		payload = payloadLocalScope
	}

	invocations++
	printMemUsage("Handler entrypoint")

	if payload.MaxPages > 0 {
		fmt.Printf("Max pages to process %v\n", payload.MaxPages)
	}

	fmt.Printf("Async saving 🔀 flag is %v\n", strings.ToUpper(strconv.FormatBool(payload.AsyncSaving)))

	if payload.SourceID == "" {
		fmt.Printf("Payload not founf or empty 💥 💥 💥 \n")
		return
	}

	heartbeatKeepAlive(payload.SourceID)
	defer heartbeatEnd()

	if !initialiseMongo() {
		return
	}
	defer closeMongo()

	printMemUsage("After mongo initialisation")

	switch payload.SourceID {
	case "GRTUKRI":
		importGrtukri()
		break
	case "PUBMED":
		importPubmed()
		break
	default:
		// NOT FOUND!
		fmt.Printf("Source ID %s not found 💥💥 \n", payload.SourceID)
		return
	}

}
