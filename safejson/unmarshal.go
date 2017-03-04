package safejson

import (
	"bytes"
	"encoding/json"
	"io"
)

func Unmarshal(data []byte, v interface{}) error {
	return UnmarshalFrom(bytes.NewReader(data), v)
}

func UnmarshalFrom(reader io.Reader, v interface{}) error {
	dec := json.NewDecoder(reader)
	dec.UseNumber()
	return dec.Decode(v)
}
