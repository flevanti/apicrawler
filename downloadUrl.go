package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer
// interface and we can pass this into io.TeeReader() which will report progress on each
// write cycle.
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func main() {
	fmt.Println("Download Started")

	fileUrl := "https://cordis.europa.eu/data/cordis-h2020projects.csv"
	err := DownloadFile("avatar.jpg", fileUrl)
	if err != nil {
		panic(err)
	}

	fmt.Println("Download Finished")
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func DownloadFile(filepath string, url string) error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	var sizeInt int
	var sizeStr string
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if sizeStr = resp.Header.Get("Content-Length"); sizeStr == "" {
		sizeStr = "0"
	}
	if sizeInt, err = strconv.Atoi(sizeStr); err != nil {
		sizeInt = 0
	}

	sizeInt = sizeInt //JUST A SILLY WAY TO USE THE VARIABLE....
	
	//TODO use size to have a real progress
	//TODO have this code encapsulated so it is possible to use it concurrently
	//TODO CREATE A WAY TO WRITE IN TMP (OUTSIDE OF THIS PIECE OF CODE...)



	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	err = os.Rename(filepath+".tmp", filepath)
	if err != nil {
		return err
	}

	return nil
}
