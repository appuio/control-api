package controllers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_invRedeemRoleName(t *testing.T) {
	require.Equal(t, "invitations-subject-redeemer", invRedeemRoleName("subject"))
	require.Equal(t,
		"invitations-subjectsubjectsubjectsubjectsubjec-c726eeb-redeemer",
		invRedeemRoleName(strings.Repeat("subject", 100)),
		"Role name must be limited to 63 characters",
	)
}
