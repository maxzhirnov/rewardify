package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHidePassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "123456789",
			input:    "host=localhost user=username password='123456789' dbname=mzhirnov sslmode=disable",
			expected: "host=localhost user=username password=******* dbname=mzhirnov sslmode=disable",
		},
		{
			name:     "cD%3d-w11Dscvwe24$$@@424352FREG4eveFW",
			input:    "host=localhost user=username password='cD%3d-w11Dscvwe24$$@@424352FREG4eveFW' dbname=mzhirnov sslmode=disable",
			expected: "host=localhost user=username password=******* dbname=mzhirnov sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := HidePassword(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
