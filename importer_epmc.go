package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type responseHeaderStructEpmc struct {
	HitCount string `json:"HitCount"`
	Request  struct {
		Query      string `json:"Query"`
		ResultType string `json:"ResultType"`
		Page       string `json:"Page"`
	} `json:"Request"`
	RecordList struct {
		Record []interface{} `json:"Record"`
	} `json:"RecordList"`
}

func importEpmc() {
	var wg sync.WaitGroup

	fmt.Printf("Endpoints parallel processing enabled? %s\n",
		strconv.FormatBool(importerConfig.EndpointsParallelProcessing))

	for _, v := range importerConfig.Endpoints {
		fmt.Printf("PROCESSING %s (Collection %s Uri %s) \n", v.Name, v.Collection, v.Uri)
		if os.Getenv("MONGO_COLLECTION_MUST_EXISTS") != "0" && !collectionExists(v.Collection) {
			fmt.Printf("Collection %s does not exists in the database! 💥 💥\n", v.Collection)
			fmt.Printf("Import skipped ❗\n")
			continue
		}
		wg.Add(1)
		if importerConfig.EndpointsParallelProcessing {
			go importEpmcEndpointLoop(v.Collection, v.Uri, &wg)
		} else {
			importEpmcEndpointLoop(v.Collection, v.Uri, &wg)
		}
	}

	fmt.Printf("WAITING FOR ALL ENDPOINTS TO BE PROCESSED.... \n")
	wg.Wait()
	fmt.Printf("DONE...\n")

}

func importEpmcEndpointLoop(collection string, uri string, wgCaller *sync.WaitGroup) {
	defer wgCaller.Done()

	currentPage := 0
	totalRecords := 0
	totalRecordsRetrieved := 0
	baseURL := ""
	client := &http.Client{}
	maxAttemptsForSameUrl := 10
	var data []byte
	var dataToSave []interface{}
	var wg sync.WaitGroup

	for {
		attempts := 0
		currentPage++
		baseURL = uri +
			"&page=" + strconv.Itoa(currentPage)

		fmt.Printf("Querying %s\n", baseURL)

		for {
			attempts++
			req, _ := http.NewRequest("GET", baseURL, nil)
			response, err := client.Do(req)

			if err != nil {
				fmt.Printf("The HTTP payload failed with error %s\n", err)
				return
			}

			data, _ = ioutil.ReadAll(response.Body)
			if response.StatusCode == 500 && attempts < maxAttemptsForSameUrl {
				fmt.Printf("The HTTP request failed (page %d) with code 500... attempt #%d/%d\n", currentPage, attempts, maxAttemptsForSameUrl)
				time.Sleep(2 * time.Second)
				continue
			}
			if response.StatusCode != 200 {
				fmt.Printf("The HTTP request failed (page %d), status code is %v\n", currentPage, response.StatusCode)
				dataString := string(data)
				if len(dataString) > 300 {
					dataString = string([]rune(dataString)[0:300])
				}
				fmt.Printf("Response received (first 300 characters....)\n%v\n", dataString)
				return
			}
			break
		}
		//remove an invalid value null and convert it to empty string
		//horrible but to the point for the moment
		data = bytes.Replace(data, []byte(":null"), []byte(":\"\""), -1)

		var dataJSON responseHeaderStructEpmc
		err := json.Unmarshal(data, &dataJSON)
		if err != nil {
			fmt.Printf("Error while converting response to json object %s 💥\n", err)
			return
		}
		dataToSave = dataJSON.RecordList.Record

		if totalRecords == 0 {
			totalRecords, _ = strconv.Atoi(dataJSON.HitCount)
		}
		totalRecordsRetrieved += len(dataToSave)
		fmt.Printf("page %d retrieved, records retrieved %d total records %d\n", currentPage, totalRecordsRetrieved, totalRecords)

		wg.Add(1)
		if importerConfig.AsyncSaving {
			go saveRecordsGrtukri(collection, dataToSave, currentPage, &wg)
		} else {
			saveRecordsGrtukri(collection, dataToSave, currentPage, &wg)
		}
		if saveRecordsErrors >= saveRecordsErrorsLimit {
			fmt.Printf("Too many errors while saving, please try again later... bye bye 👋 ...\n")
			break
		}

		if importerConfig.MaxPages > 0 && currentPage >= importerConfig.MaxPages {
			fmt.Printf("All pages retrieved...\n")
			break
		}

		if totalRecordsRetrieved >= totalRecords {
			fmt.Printf("All pages retrieved...\n")
			break
		}

		printMemUsage("while looping...")
	} //end for loop

	printMemUsage("Loops just ended...")
	fmt.Printf("Wait for all processes to complete.... %s\n", getDT())
	wg.Wait()
	fmt.Printf("Done! %s\n", getDT())
	printMemUsage("Wait is over!")
}
