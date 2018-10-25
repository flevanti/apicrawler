package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type responseHeaderStructGrtukri struct {
	Organisation []interface{} `json:"organisation"`
	Fund         []interface{} `json:"fund"`
	Person       []interface{} `json:"person"`
	Project      []interface{} `json:"project"`
	Page         int           `json:"page"`
	Size         int           `json:"size"`
	TotalPages   int           `json:"totalPages"`
	TotalSize    int           `json:"totalSize"`
}

func importGrtukri() {
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
			go importGrtukriEndpointLoop(v.Name, v.Collection, v.Uri, v.ResponseElement, &wg)
		} else {
			importGrtukriEndpointLoop(v.Name, v.Collection, v.Uri, v.ResponseElement, &wg)
		}
	}

	fmt.Printf("WAITING FOR ALL ENDPOINTS TO BE PROCESSED.... \n")
	wg.Wait()
	fmt.Printf("DONE...\n")

}

func importGrtukriEndpointLoop(name string, collection string, uri string, responseElement string, wgCaller *sync.WaitGroup) {
	defer wgCaller.Done()

	currentPage := 0
	baseURL := ""
	size := importerConfig.PageSize
	client := &http.Client{}
	var dataToSave []interface{}
	var wg sync.WaitGroup

	for {
		currentPage++
		baseURL = uri +
			"?s=" + strconv.Itoa(size) +
			"&p=" + strconv.Itoa(currentPage)
		fmt.Printf("Querying %s\n", baseURL)

		req, _ := http.NewRequest("GET", baseURL, nil)
		req.Header.Set("Accept", "application/vnd.rcuk.gtr.json-v6")
		response, err := client.Do(req)

		if err != nil {
			fmt.Printf("The HTTP payload failed with error %s\n", err)
			return
		}

		data, _ := ioutil.ReadAll(response.Body)
		if response.StatusCode != 200 {
			fmt.Printf("The HTTP payload failed, status code is %v\n", response.StatusCode)
			dataString := string(data)
			if len(dataString) > 300 {
				dataString = string([]rune(dataString)[0:300])
			}
			fmt.Printf("Response received (first 300 characters....)\n%v\n", dataString)
			return
		}

		var dataJSON responseHeaderStructGrtukri
		err = json.Unmarshal([]byte(data), &dataJSON)
		if err != nil {
			fmt.Printf("Error while converting response to json object %s 💥\n", err)
			return
		}
		fmt.Printf("page %d retrieved, size %d, total pages %d, total size %d\n", dataJSON.Page, dataJSON.Size, dataJSON.TotalPages, dataJSON.TotalSize)

		//check what data to save....
		switch responseElement {
		case "organisation":
			dataToSave = dataJSON.Organisation
			break
		case "fund":
			dataToSave = dataJSON.Fund
			break
		case "person":
			dataToSave = dataJSON.Person
			break
		case "project":
			dataToSave = dataJSON.Project
			break
		default:
			dataToSave = nil
		}

		if dataToSave == nil {
			fmt.Printf("Unable to find records extracted in element  😒 ...\n")
			saveRecordsErrors++

		}

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

		if currentPage == dataJSON.TotalPages {
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
