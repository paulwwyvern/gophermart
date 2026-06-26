package luhn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{
			name: "Test #1 Valid",
			s:    "12345678903",
			want: 0,
		},
		{
			name: "Test #2 Valid",
			s:    "2436",
			want: 0,
		},
		{
			name: "Test #3 Invalid",
			s:    "123456789",
			want: 1,
		},
		{
			name: "Test #4 Invalid",
			s:    "1",
			want: 1,
		},
		{
			name: "Test #5 Invalid",
			s:    "asdf",
			want: -1,
		},
		{
			name: "Test #6 Invalid",
			s:    "",
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Validate(tt.s))
		})
	}
}
