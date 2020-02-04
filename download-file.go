//////
// Credit to Golang code for their awesome work [golangcode.com]
// Credit to Dustin for Humanize [github.com/dustin/go-humanize]
//////
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
)

// WriteCounter counts the number of bytes written to it.  It creates an io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type WriteCounter struct {
	Written uint64
}

var fileContentLength string

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Written += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces.
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download.
	fmt.Printf("\rDownloading: %s/%s complete", humanize.Bytes(wc.Written), fileContentLength)
}

func main() {
	fileURL := "https://www.wwf.org.au/Images/UserUploadedImages/416/img-koala-eating-leaf-queensland-1000px.jpg"
	fileName := stringAfter(fileURL, "/")

	fmt.Println("Started downloading", fileURL)

	err := DownloadFile(fileName, fileURL)
	if err != nil {
		panic(err)
	}

	fmt.Println("Finished downloading", fileName)
}

func stringAfter(value string, a string) string {
	// Get substring after a string/character.
	// Retrieves the last instance of substring "a"
	position := strings.LastIndex(value, a)
	if position == -1 {
		return ""
	}
	adjustedPosition := position + len(a)
	if adjustedPosition >= len(value) {
		return ""
	}
	return value[adjustedPosition:len(value)]
}

func DownloadFile(filepath string, url string) error {
	// Downloads file from url to a local file.
	// It writes as it downloads and does not load the whole file to memory.
	// io.TeeReader in Copy reports progress of the download.

	// Create file with tmp file extension.
	// Remove the tmp extension once downloaded.
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return err
	}

	// Get data
	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}
	i, err := humanize.ParseBytes(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	fileContentLength = humanize.Bytes(i)
	defer resp.Body.Close()

	// Create progress reporter and pass it to be used alongside writer.
	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	// The progress reporter uses the same line \
	// so print a new line one download is finished.
	fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return err
	}
	return nil
}
