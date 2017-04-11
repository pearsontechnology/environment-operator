package bitesize

import (
	"fmt"
	"testing"
)

func TestValidationVolumeNames(t *testing.T) {
	var testCases = []struct {
		Value interface{}
		Error error
	}{
		{
			"redbluegreen",
			fmt.Errorf("Invalid volume mode: redbluegreen"),
		},
		{
			1,
			fmt.Errorf("Invalid volume mode: 1. Valid modes: ReadWriteOnce,ReadOnlyMany,ReadWriteMany"),
		},
		{
			"ReadWriteOnce",
			nil,
		},
	}

	for _, tCase := range testCases {
		err := validVolumeModes(tCase.Value, "")
		if err != tCase.Error {
			if err.Error() != tCase.Error.Error() {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	}

}
