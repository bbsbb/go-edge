package datatype

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func ScanJSON[T any](ptr *T, value any) error {
	switch v := value.(type) {
	case []byte:
		dec := json.NewDecoder(bytes.NewReader(v))
		dec.DisallowUnknownFields()
		return dec.Decode(ptr)
	case string:
		dec := json.NewDecoder(bytes.NewReader([]byte(v)))
		dec.DisallowUnknownFields()
		return dec.Decode(ptr)
	default:
		return fmt.Errorf("invalid value type '%T' for JSON scan", value)
	}
}
