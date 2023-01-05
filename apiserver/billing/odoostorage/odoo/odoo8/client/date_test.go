package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		givenInput   string
		expectedDate *Date
	}{
		"GivenFalse_ThenExpectZeroDate": {
			givenInput:   "false",
			expectedDate: nil,
		},
		"GivenValidInput_WhenFormatIsDate_ThenExpectDate": {
			givenInput:   "2021-02-03",
			expectedDate: newDate(t, "2021-02-03"),
		},
		"GivenValidInput_WhenFormatIsDateTime_ThenExpectDateTime": {
			givenInput:   "2021-02-03 15:34:00",
			expectedDate: newDateTime(t, "2021-02-03 15:34"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subject := &Date{}
			err := subject.UnmarshalJSON([]byte(tt.givenInput))
			require.NoError(t, err)
			if tt.expectedDate == nil {
				assert.True(t, subject.IsZero())
				return
			}
			assert.Equal(t, tt.expectedDate, subject)
		})
	}
}

func newDateTime(t *testing.T, value string) *Date {
	tm, err := time.Parse(DateTimeFormat, fmt.Sprintf("%s:00", value))
	require.NoError(t, err)
	ptr := Date(tm)
	return &ptr
}

func newDate(t *testing.T, value string) *Date {
	tm, err := time.Parse(DateFormat, value)
	require.NoError(t, err)
	ptr := Date(tm)
	return &ptr
}
