package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOTPGenerationAndValidation(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, secret)

	uri := GenerateTOTPURI(secret, "FileBrowser", "admin")
	assert.Contains(t, uri, "otpauth://totp/FileBrowser:admin")
	assert.Contains(t, uri, secret)

	// Current code should be valid
	currentCode, err := GenerateCode(secret, time.Now())
	require.NoError(t, err)
	assert.Len(t, currentCode, 6)

	assert.True(t, ValidateTOTP(currentCode, secret))

	// Invalid code should fail
	assert.False(t, ValidateTOTP("000000", secret))
	assert.False(t, ValidateTOTP("invalid", secret))
}
