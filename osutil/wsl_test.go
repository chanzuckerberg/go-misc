package osutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWSLRegex(t *testing.T) {
	r := require.New(t)

	type test struct {
		in       string
		expected bool
	}

	tests := []test{
		{in: "microsoft", expected: true},
		{in: "WSL", expected: true},
		{in: "Microsoft", expected: true},
		{in: "does not match", expected: false},
	}

	for _, test := range tests {
		r.Equal(test.expected, reWSL.MatchString(test.in))
	}
}
