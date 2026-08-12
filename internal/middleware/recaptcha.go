package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const recaptchaVerifyURL = "https://www.google.com/recaptcha/api/siteverify"

type recaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

// VerifyRecaptcha validates a Google reCAPTCHA v2 token against Google's siteverify API.
// Returns nil on success, error on failure.
func VerifyRecaptcha(secretKey, token, remoteIP string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("reCAPTCHA token wajib diisi")
	}

	payload := fmt.Sprintf("secret=%s&response=%s&remoteip=%s", secretKey, token, remoteIP)

	resp, err := http.Post(
		recaptchaVerifyURL,
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(payload),
	)
	if err != nil {
		return fmt.Errorf("gagal menghubungi server reCAPTCHA: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gagal membaca respons reCAPTCHA")
	}

	var result recaptchaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("respons reCAPTCHA tidak valid")
	}

	if !result.Success {
		codes := strings.Join(result.ErrorCodes, ", ")
		return fmt.Errorf("verifikasi reCAPTCHA gagal: %s", codes)
	}

	return nil
}
