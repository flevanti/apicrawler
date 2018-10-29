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

type responseHeaderStructOpenaire struct {
	Response struct {
		Header struct {
			Query struct {
				Query string `json:"$"`
			} `json:"query"`
			Locale struct {
				Locale string `json:"$"`
			} `json:"locale"`
			Size struct {
				Size int `json:"$"`
			} `json:"size"`
			Page struct {
				Page int `json:"$"`
			} `json:"page"`
			Total struct {
				Total int `json:"$"`
			} `json:"total"`
			Fields interface{} `json:"fields"`
		} `json:"header"`
		Results struct {
			Result []interface{} `json:"result"`
		} `json:"results"`
	} `json:"response"`
}

func importOpenaire() {
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
			go importOpenaireEndpointLoop(v.Collection, v.Uri, &wg)
		} else {
			importOpenaireEndpointLoop(v.Collection, v.Uri, &wg)
		}
	}

	fmt.Printf("WAITING FOR ALL ENDPOINTS TO BE PROCESSED.... \n")
	wg.Wait()
	fmt.Printf("DONE...\n")

}

func importOpenaireEndpointLoop(collection string, uri string, wgCaller *sync.WaitGroup) {
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
			"&page=" + strconv.Itoa(currentPage) + "&size=" + strconv.Itoa(importerConfig.PageSize)

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
		//
		// PLEASE NOTE the space between colon and the null string
		data = bytes.Replace(data, []byte(": null"), []byte(": \"\""), -1)
		// the generated json does not respect some basic rules like...
		// a number can be NOT wrapped with quotes but cannot start with leading zeros
		// golang json unmarshaller is very fussy, yes....
		// one more of these crappy situations and I'll switch to XML
		// so if I have to cry.. I want to cry for a reason....
		data = bytes.Replace(data, []byte(": 00"), []byte(": "), -1)

		var dataJSON responseHeaderStructOpenaire
		err := json.Unmarshal(data, &dataJSON)
		if err != nil {
			fmt.Printf("Error while converting response to json object %s 💥\n", err)
			return
		}
		dataToSave = dataJSON.Response.Results.Result

		if totalRecords == 0 {
			totalRecords = dataJSON.Response.Header.Total.Total
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
