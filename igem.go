package main

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	dbPath            = "db"
	allPartURL        = "http://parts.igem.org/fasta/parts/All_Parts"
	allPartFilename   = "allpart.dump"
	downloadBlockSize = 1 << 13 // 8k, size of a packet
)

type Biobrick struct {
	name     string
	letter   rune
	number   int
	desc     string
	sequence string
}

func main() {

	var partString string

	for i, val := range strings.Split(partString, ">") {
		fmt.Println(val)
		if i > 10 {
			return
		}
	}
}

func GetAllPartsString() (string, error) {
	partRaw, err := ioutil.ReadFile(allPartFilename)
	if err != nil {
		partString, err := DownloadAllParts()
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(allPartFilename, []byte(partString), 0775)
		if err != nil {
			log.Printf("Failed to save %s file, %v\n", allPartFilename, err)
		}
		return partString, nil
	}

	return string(partRaw), nil

}

func DownloadAllParts() (string, error) {

	resp, err := http.Get(allPartURL)
	if err != nil {
		return "", fmt.Errorf("getting %s, %v", allPartURL, err)
	}
	defer resp.Body.Close()

	progressUpdt := GetProgressFunc(resp.ContentLength)

	fileReader, err := DownloadFile(resp.Body, resp.ContentLength, progressUpdt)
	if err != nil {
		return "", fmt.Errorf("reading content from body, %v", err)
	}

	node, err := html.Parse(fileReader)
	if err != nil {
		return "", fmt.Errorf("parsing file response, %v", err)
	}

	partContent := goquery.NewDocumentFromNode(node).Find("pre")

	if partContent == nil {
		return "", fmt.Errorf("expected content but none found")
	}
	return partContent.Text(), nil
}

func GetProgressFunc(total int64) func(int64) {
	return func(i int64) {
		percDone := float64(i) / float64(total) * 100.0
		fmt.Printf("%3.2f percent done, %d/%d bytes\r", percDone, i, total)
	}
}

func DownloadFile(r io.Reader, totalSize int64, progressUpdt func(i int64)) (io.Reader, error) {
	out := bytes.NewBuffer(make([]byte, 0, totalSize))
	byteRead := int64(0)

	fmt.Println("Download starts")
	for {

		n, err := io.CopyN(out, r, downloadBlockSize)

		byteRead += n
		progressUpdt(byteRead)

		if n < downloadBlockSize {
			break
		} else if err != nil {
			return nil, err
		}

	}
	fmt.Println("\nDownload done")

	return out, nil
}
