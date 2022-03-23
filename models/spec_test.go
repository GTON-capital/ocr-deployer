package models

import (
	"testing"
)

func TestParseSpec(t *testing.T) {

	got, err := LoadSpec("/home/alex/development/ocr-deployer/specs/ocr-PEG-Base.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got.Description != "OCR test" {
		t.Log("Invalid value")
		t.FailNow()
	}

}
