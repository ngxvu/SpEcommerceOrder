package ginext

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	raw := []byte(`"2021-07-18"`)
	date := JsDate{}

	err := json.Unmarshal(raw, &date)
	assert.NoError(t, err)

	tDate := time.Time(date)
	assert.Equal(t, 2021, tDate.Year())
	assert.Equal(t, 7, int(tDate.Month()))
	assert.Equal(t, 18, tDate.Day())
}

func TestJsDateMarshal(t *testing.T)  {
	tDate, err := time.Parse(DateLayout, "2021-07-18")
	require.NoError(t, err)
	date := JsDate(tDate)
	s, _ := json.Marshal(date)
	assert.Equal(t, `"2021-07-18"`, string(s))
}
