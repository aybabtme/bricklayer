package bricks

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestExtendedBiobrick(t *testing.T) {
	parts, err := QueryExtendedBiobricks("BBa_B0034")
	if err != nil {
		t.Fatalf("%v", err)
	}

	fmt.Printf("%#v\n", parts)

	data, err := json.Marshal(parts)
	if err != nil {
		t.Fatalf("%v", err)
	}

	fmt.Printf("\n\n%s\n", string(data))
}
