package utils

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"net/url"
)

func EncodeURL(urlStr string) string {
	// URL encoding
	encodedURL := url.QueryEscape(urlStr)
	// Base64 encoding
	encodedURL = base64.URLEncoding.EncodeToString([]byte(encodedURL))
	return encodedURL
}

// EncodeAndCompressURL encodes and compresses the given URL.
func EncodeAndCompressURL(urlString string) (string, error) {
	// Encode the URL
	encodedURL := url.QueryEscape(urlString)

	// Compress the encoded URL
	var compressedURL bytes.Buffer
	compressor := zlib.NewWriter(&compressedURL)
	_, err := compressor.Write([]byte(encodedURL))
	if err != nil {
		return "", fmt.Errorf("error compressing URL: %w", err)
	}
	compressor.Close()

	return compressedURL.String(), nil
}

// DecompressAndDecodeURL decompresses and decodes the given compressed URL.
func DecompressAndDecodeURL(compressedURL string) (string, error) {
	// Convert the compressed URL from a string to a byte slice
	compressedBytes, err := base64.StdEncoding.DecodeString(compressedURL)
	if err != nil {
		return "", fmt.Errorf("error decoding base64 URL: %w", err)
	}

	// Decompress the compressed URL
	decompressedURL, err := zlib.NewReader(bytes.NewReader(compressedBytes))
	if err != nil {
		return "", fmt.Errorf("error decompressing URL: %w", err)
	}
	defer decompressedURL.Close()

	// Decode the URL
	var decompressedBuffer bytes.Buffer
	_, err = decompressedBuffer.ReadFrom(decompressedURL)
	if err != nil {
		return "", fmt.Errorf("error reading decompressed URL: %w", err)
	}

	decodedURL, err := url.QueryUnescape(decompressedBuffer.String())
	if err != nil {
		return "", fmt.Errorf("error decoding URL: %w", err)
	}

	return decodedURL, nil
}
