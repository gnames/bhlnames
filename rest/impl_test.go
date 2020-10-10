package rest

import (
	"fmt"
	"testing"

	linkent "github.com/gdower/bhlinker/domain/entity"
)

func TestFixIDs(t *testing.T) {
	input := []linkent.Input{
		{},
		{ID: "1"},
		{ID: "1"},
		{ID: "uniq"},
	}
	fixIDs(input)
	if len(input[0].ID) < 10 || input[1].ID != "1" || len(input[2].ID) < 10 || input[3].ID != "uniq" {
		fmt.Printf("\nfixIDs: %+v\n\n", input)
		t.Error("IDs did not change correctly")
	}
}
