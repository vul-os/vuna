package utils

import (
	"encoding/base64"
	"net/url"
)

func EncodeURL(urlStr string) string {
	// URL encoding
	encodedURL := url.QueryEscape(urlStr)
	// Base64 encoding
	encodedURL = base64.URLEncoding.EncodeToString([]byte(encodedURL))
	return encodedURL
}