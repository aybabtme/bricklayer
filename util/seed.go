package util

import (
	"fmt"
	"github.com/aybabtme/bricklayer/bricks"
	"net/http"
)

const (
	dbPath          = "db"
	allPartFilename = "allpart.dump"

	iGemAllPartURL = "http://parts.igem.org/fasta/parts/All_Parts"
	// iGemAllPartURL = "http://127.0.0.1:8080/All_Parts"

	downloadBlockSize = 1 << 13 // 8k, size of a packet
)

func DownloadAllBiobricks() ([]bricks.Biobrick, error) {

	resp, err := http.Get(iGemAllPartURL)
	if err != nil {
		return nil, fmt.Errorf("getting %s, %v", iGemAllPartURL, err)
	}
	defer resp.Body.Close()

	bioReader := bricks.NewBiobrickReader(resp.Body)
	count := 0

	var allBioBricks []bricks.Biobrick
	for {

		biobrick, err := bioReader.Read()
		if err != nil {
			return nil, err
		}
		if biobrick != nil {
			count++
			fmt.Printf("Reading bricks : %d parts found\r", count)
			allBioBricks = append(allBioBricks, *biobrick)
		} else {
			fmt.Println("")
			break
		}
	}
	fmt.Println("Done")

	return allBioBricks, nil
}
