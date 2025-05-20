package ginext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagerGetOrder(t *testing.T) {
	cases := []struct {
		name       string
		inputOrder string
		output     string
	}{
		{
			name:       "EmptyByDefault",
			inputOrder: "",
			output:     "",
		},
		{
			name:       "ConvertMinusToDesc",
			inputOrder: "id,-name",
			output:     "id asc, name desc",
		},
		{
			name:       "TrimSpaces",
			inputOrder: "id,  -name",
			output:     "id asc, name desc",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Pager{
				SortableFields: []string{"id", "name"}, // allow sort 2 these fields by default
				Sort: tc.inputOrder,
			}
			got := p.GetOrder(p.SortableFields)
			assert.Equal(t, tc.output, got)
		})
	}
}
