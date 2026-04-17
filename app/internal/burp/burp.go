package burp

import (
	"bytes"
	"encoding/xml"

	"golang.org/x/net/html/charset"
)

func Parse(content []byte) (*BurpData, error) {
	dec := xml.NewDecoder(bytes.NewReader(content))
	dec.CharsetReader = charset.NewReaderLabel
	dec.Strict = false

	var doc BurpData
	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}

	return &doc, nil
}
