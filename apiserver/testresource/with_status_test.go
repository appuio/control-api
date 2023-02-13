package testresource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	subject := &TestResourceWithStatus{}
	(&TestResourceWithStatusStatus{Num: 7}).SecretStorageCopyTo(subject)

	require.Equal(t, 7, subject.SecretStorageGetStatus().(*TestResourceWithStatusStatus).Num)
}
