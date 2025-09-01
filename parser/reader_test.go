package parser

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestArrayString(t *testing.T) {
	testCases := map[string]struct {
		has  string
		want []string
	}{
		"correctly parse PING": {
			has:  "1\r\n$4\r\nPING\r\n",
			want: []string{"PING"},
		},
		"correctly parse ECHO": {
			has:  "2\r\n$4\r\nECHO\r\n$3\r\nHEY\r\n",
			want: []string{"ECHO", "HEY"},
		},
		"correctly parse GET": {
			has:  "2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n",
			want: []string{"GET", "FOO"},
		},
		"correctly parse SET": {
			has:  "3\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n",
			want: []string{"SET", "FOO", "BAR"},
		},
		"correctly parse SET with ttl": {
			has:  "5\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n$2\r\nPX\r\n$3\r\n100\r\n",
			want: []string{"SET", "FOO", "BAR", "PX", "100"},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			br := bufio.NewReader(strings.NewReader(tc.has))
			reader := NewReader(br)
			resp, err := reader.ArrayString()
			require.NoError(t, err)
			assert.Equal(t, tc.want, resp)
		})
	}
}
