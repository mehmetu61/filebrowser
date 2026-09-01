package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultTOTPPeriod is standard 30-second window
	DefaultTOTPPeriod = 30
	// DefaultTOTPDigits is standard 6 digits
	DefaultTOTPDigits = 6
)

// GenerateTOTPSecret creates a random base32 encoded secret (20 bytes / 160 bits).
func GenerateTOTPSecret() (string, error) {
	secretBytes := make([]byte, 20)
	_, err := rand.Read(secretBytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes), nil
}

// GenerateTOTPURI returns the otpauth:// URI for scanning with Authenticator apps.
func GenerateTOTPURI(secret, issuer, username string) string {
	encodedIssuer := url.QueryEscape(issuer)
	encodedUsername := url.QueryEscape(username)
	cleanSecret := strings.ToUpper(strings.TrimSpace(secret))
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		encodedIssuer, encodedUsername, cleanSecret, encodedIssuer, DefaultTOTPDigits, DefaultTOTPeriod)
}

// GenerateCode calculates the TOTP code for a given timestamp.
func GenerateCode(secret string, t time.Time) (string, error) {
	cleanSecret := strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleanSecret)
	if err != nil {
		// Try with standard padding
		key, err = base32.StdEncoding.DecodeString(cleanSecret)
		if err != nil {
			return "", fmt.Errorf("invalid base32 secret: %w", err)
		}
	}

	counter := uint64(t.Unix() / DefaultTOTPPeriod)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff
	code := truncated % 1000000

	return fmt.Sprintf("%06d", code), nil
}

// ValidateTOTP checks if the given passcode is valid within a window of [-1, +1] time steps.
func ValidateTOTP(passcode, secret string) bool {
	cleanPasscode := strings.TrimSpace(passcode)
	if len(cleanPasscode) != DefaultTOTPDigits {
		return false
	}

	now := time.Now()
	// Check previous, current, and next time windows to account for clock skew
	for _, offset := range []int64{-1, 0, 1} {
		t := now.Add(time.Duration(offset*DefaultTOTPPeriod) * time.Second)
		code, err := GenerateCode(secret, t)
		if err != nil {
			continue
		}
		if code == cleanPasscode {
			return true
		}
	}
	return false
}
