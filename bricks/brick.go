package bricks

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Biobrick struct {
	bioAbstract
	Sequence string `json: "sequence"`
}

type bioAbstract struct {
	PartName string `json: "partName"`
	Status   string `json: "status"`
	ID       int    `json: "id"`
	Type     string `json: "type"`
	Desc     string `json: "desc"`
}

func abstractFromString(abstract string) (*bioAbstract, error) {
	parts := strings.Split(abstract, " ")

	if len(parts) < 5 {
		return nil, fmt.Errorf("got %d instead of 5 expected parts in abstract '%s'", len(parts), abstract)
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("need integer for biopart ID, %v", err)
	}
	descWithQuote := strings.Join(parts[4:], " ")
	return &bioAbstract{
		PartName: parts[0],
		Status:   parts[1],
		ID:       id,
		Type:     parts[3],
		Desc:     descWithQuote[1 : len(descWithQuote)-1], // Remove quotes
	}, err
}

type BiobrickReader struct {
	scanner *bufio.Scanner
}

func NewBiobrickReader(r io.Reader) *BiobrickReader {

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	return &BiobrickReader{scanner: scanner}
}

const (
	prefix = ">"
)

func (b *BiobrickReader) Read() (*Biobrick, error) {

	if !b.scanner.Scan() {
		return nil, b.scanner.Err()
	}

	text := b.scanner.Text()
	if !strings.HasPrefix(text, prefix) {
		return nil, fmt.Errorf("expect prefix for FASTA format")
	}

	abstract, err := abstractFromString(text[1:])
	if err != nil {
		return nil, fmt.Errorf("reading abstract for biobrick, %v", err)
	}

	descBuf := bytes.NewBuffer(nil)
	for b.scanner.Scan() {
		descBytes := b.scanner.Bytes()
		if len(descBytes) == 0 {
			break
		}

		n, err := descBuf.Write(descBytes)
		if err != nil {
			return nil, fmt.Errorf("reading line from description, %v", err)
		}
		if n != len(descBytes) {
			return nil, fmt.Errorf("should have written %d bytes but wrote %d", len(descBytes), n)
		}
	}

	return &Biobrick{*abstract, descBuf.String()}, b.scanner.Err()
}
