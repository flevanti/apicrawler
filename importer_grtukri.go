package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func importGRTUKRI() {

	currentPage := 0
	baseURL := ""
	size := payload.PageSize

	if !initialiseMongo() {
		return
	}
	defer closeMongo()

	fmt.Printf("Mongo collection is %s\n", mongoCollection.Name())
	printMemUsage("After mongo initialisation")

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
			fmt.Printf("The HTTP payload failed with error %s\n", err)
			return
		}
		defer response.Body.Close()
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

		var dataJSON responseHeaderStruct
		err = json.Unmarshal([]byte(data), &dataJSON)
		if err != nil {
			fmt.Printf("Error while converting response to json object %s 💥\n", err)
			return
		}
		fmt.Printf("page %d retrieved, size %d, total pages %d, total size %d\n", dataJSON.Page, dataJSON.Size, dataJSON.TotalPages, dataJSON.TotalSize)

		wg.Add(1)
		if payload.AsyncSaving {
			go saveRecords(dataJSON, currentPage)
		} else {
			saveRecords(dataJSON, currentPage)
		}

		if payload.MaxPages > 0 && payload.MaxPages == currentPage {
			fmt.Printf("All pages retrieved...\n")
			break
		}

		if currentPage == dataJSON.TotalPages {
			fmt.Printf("All pages retrieved...\n")
			break
		}

		if saveRecordsErrors >= saveRecordsErrorsLimit {
			fmt.Printf("Too many errors while saving, please try again later... bye bye 👋🏼...\n")
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
