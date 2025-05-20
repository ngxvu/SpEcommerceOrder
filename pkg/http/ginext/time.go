package ginext

import (
	"bytes"
	"time"
)

const (
	DateLayout = "2006-01-02"
)

type JsDate time.Time

func (j *JsDate) UnmarshalJSON(b []byte) error {
	// get rid of "
	value := string(bytes.Trim(b, `"`))
	if value == "" || value == "null" {
		return nil
	}
	t, err := time.Parse(DateLayout, value)
	if err != nil {
		return err
	}
	*j = JsDate(t) // set result using pointer
	return nil
}

func (j JsDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(j).Format(DateLayout) + `"`), nil
}
