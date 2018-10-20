package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var a int
var isLambda bool

type responseHeaderStruct struct {
	Organisation []interface{}
	Page         int `json:"page"`
	Size         int `json:"size"`
	TotalPages   int `json:"totalPages"`
	TotalSize    int `json:"totalSize"`
}

var wg sync.WaitGroup

var mongoClient *mongo.Client
var mongoDb *mongo.Database
var mongoCollection *mongo.Collection

var maxPages = flag.Int("maxpages", 0, "max number of pages retrieved (for testing?)")

func main() {
	PrintMemUsage("main entrypoint 🍾")
	flag.Parse()
	if !loadEnvVariables() {
		fmt.Printf("Unable to read .ENV file 💥 \n")
		return
	}
	checkEnvironment()
	greetings()
	fmt.Printf("We may think I'm migrating this from old languages to improve stability and performance\n")
	fmt.Printf("The truth is I just wanted support for emojis.... 🚀 🤠 🍍\n")
	if isLambda {
		lambda.Start(Handler)
	} else {
		Handler()
	}
	PrintMemUsage("main exit 🍾 🕺🏼 🍾")

}

// Handler doc block....
func Handler() {
	PrintMemUsage("Handler entrypoint")
	if *maxPages > 0 {
		fmt.Printf("Max pages to process %v\n", *maxPages)
	}

	currentPage := 0
	baseURL := ""
	/*
		baseFilename := "GTRUKRI_ORGS"
		filename := ""
		filenameExtension := ".json"
	*/
	size := 50

	initialiseMongo()
	fmt.Printf("Mongo collection is %s\n", mongoCollection.Name())
	PrintMemUsage("After mongo initialisation")

	for {
		currentPage++
		baseURL = "https://gtr.ukri.org/gtr/api/organisations?s=" +
			strconv.Itoa(size) + "&p=" + strconv.Itoa(currentPage)
		fmt.Printf("Querying %s\n", baseURL)
		client := &http.Client{}
		req, _ := http.NewRequest("GET", baseURL, nil)
		req.Header.Set("Accept", "application/vnd.rcuk.gtr.json-v6")
		response, err := client.Do(req)

		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
			return
		}
		defer response.Body.Close()
		data, _ := ioutil.ReadAll(response.Body)
		if response.StatusCode != 200 {
			fmt.Printf("The HTTP request failed, status code is %v\n", response.StatusCode)
			dataString := string(data)
			if len(dataString) > 300 {
				dataString = string([]rune(dataString)[0:300])
			}
			fmt.Printf("Response received (first 300 characters....)\n%v\n", dataString)
			return
		}

		var dataJSON responseHeaderStruct
		err = json.Unmarshal([]byte(data), &dataJSON)
		if err != nil {
			fmt.Printf("Error while converting response to json object %s 💥\n", err)
			return
		}
		fmt.Printf("page %d retrieved, size %d, total pages %d, total size %d\n", dataJSON.Page, dataJSON.Size, dataJSON.TotalPages, dataJSON.TotalSize)

		/*
			filename = baseFilename +
				"_p" + strconv.Itoa(currentPage) +
				"_s" + strconv.Itoa(size) +
				"_tp" + strconv.Itoa(dataJSON.TotalPages) +
				filenameExtension
			ioutil.WriteFile(filename, data, 0644)
			fmt.Printf("Filename created %s\n", filename)

			time.Sleep(500 * time.Millisecond)
		*/
		wg.Add(1)
		go saveRecords(dataJSON, currentPage)

		if *maxPages > 0 && *maxPages == currentPage {
			fmt.Printf("All pages retrieved...\n")
			break
		}

		if currentPage == dataJSON.TotalPages {
			fmt.Printf("All pages retrieved...\n")
			break
		}
		PrintMemUsage("while looping...")
	} //end for loop

	PrintMemUsage("Loops just ended...")
	fmt.Printf("Wait for all processes to complete.... %s\n", getDT())
	wg.Wait()
	fmt.Printf("Done! %s\n", getDT())
	PrintMemUsage("Wait is over!")

}

func saveRecords(data responseHeaderStruct, page int) {
	defer wg.Done()
	fmt.Printf("Request to save page %d ...\n", page)
	_, err := mongoCollection.InsertMany(context.Background(), data.Organisation)
	if err != nil {
		fmt.Printf("Error while saving to mongo: %s\n", err)
		return
	}
	fmt.Printf("Page %d saved!\n", page)

}

// checkEnvironment
func checkEnvironment() {
	if len(os.Getenv("AWS_REGION")) != 0 {
		isLambda = true
	} else {
		isLambda = false
	}
}

func greetings() {
	if isLambda {
		fmt.Printf("Hello Jeff\n")
	} else {
		fmt.Printf("Greetings Professor Falken")
	}
}

func getDT() string {
	// time date formatting...
	// https://golang.org/src/time/format.go
	return time.Now().Format("2006-01-02 15:04:05.0000")
}

func loadEnvVariables() bool {
	//read .env variables
	if err := godotenv.Load(); err != nil {
		return false
	}
	return true
}

func initialiseMongo() {
	mongoClient, _ = mongo.Connect(context.Background(), os.Getenv("MONGO_URI"), nil)
	mongoClient.Connect(context.Background())
	mongoDb = mongoClient.Database(os.Getenv("MONGO_DATABASE"))
	mongoCollection = mongoDb.Collection("MONGO_COLLECTION")
}

// PrintMemUsage
func PrintMemUsage(context string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("MiB usage malloc %v  tot malloc %v  Sys %v  GC %v  (%s)\n",
		b2KiB(m.Alloc), b2KiB(m.TotalAlloc), b2KiB(m.Sys), m.NumGC, context)
}

func b2KiB(b uint64) uint64 {
	return b / 1024
}
