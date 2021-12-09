package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUrlMap_GetURL(t *testing.T) {
	tests := map[string]struct {
		givenMap      URLMap
		givenKey      string
		expectedErr   string
		expectedValue string
	}{
		"GivenNil_ThenExpectError": {
			expectedErr: "map is nil",
		},
		"GivenEmptyMap_WhenKeyNotPresent_ThenExpectError": {
			givenMap:    URLMap{},
			givenKey:    "key",
			expectedErr: "key not found: key",
		},
		"GivenMap_WhenKeyPresent_ThenParseUrl": {
			givenMap: URLMap{
				"key": "https://hostname:80/path",
			},
			givenKey:      "key",
			expectedValue: "https://hostname:80/path",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := tt.givenMap.GetURL(tt.givenKey)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
				return
			}
			assert.Equal(t, tt.expectedValue, result.String())
		})
	}
}
