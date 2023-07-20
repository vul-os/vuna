package utils

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const BACKSLACK_REP = "&($)!"
const EQUALS_REP = "&$$&"
const CHAR_COMBO = "@{&$!"

func GetHostName(rawURL string) (string, string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("error parsing URL: %v", err)
	}
	hostId := RemoveWWWPrefix(RemoveHTTPPrefix(parsedURL.Host))
	productId := RemoveWWWPrefix(RemoveHTTPPrefix(parsedURL.RawPath))
	return hostId, productId, nil
}

func RemoveHTTPPrefix(input string) string {
	re := regexp.MustCompile(`^(http://|https://)`)
	output := re.ReplaceAllString(input, "")
	return output
}

func RemoveWWWPrefix(input string) string {
	re := regexp.MustCompile(`^www\.`)
	output := re.ReplaceAllString(input, "")
	return output
}

func StringToSiteIdentifier(urlString string, otherStringIds []string) (string, string, error) {
	// Fixed character combination
	hostString, _, err := GetHostName(urlString)

	if err != nil {
		return "", "", err
	}

	return StringToXIdentifier(hostString, otherStringIds)
}

func StringToProductIdentifier(urlString string, otherStringIds []string) (string, string, error) {
	// Fixed character combination
	_, productString, err := GetHostName(urlString)

	if err != nil {
		return "", "", err
	}
	return StringToXIdentifier(productString, otherStringIds)
}

func StringToXIdentifier(stringy string, otherStringIds []string) (string, string, error) {
	// Concatenate the strings with the fixed character combination
	stringsToJoin := append(otherStringIds, stringy)
	joinedString := strings.Join(stringsToJoin, CHAR_COMBO)
	encoded4URL := url.QueryEscape(joinedString)

	encodedString, err := EncodeAndCompressString(encoded4URL)
	if err != nil {
		return "", "", err
	}

	encodedHostString, err := EncodeAndCompressString(stringy)
	if err != nil {
		return "", "", err
	}
	return encodedHostString, encodedString, nil
}

func EncodeURL(urlStr string) string {
	// URL encoding
	encodedURL := url.QueryEscape(urlStr)
	// Base64 encoding
	encodedURL = base64.URLEncoding.EncodeToString([]byte(encodedURL))
	return encodedURL
}

func replacePaddingChars(encodedString string) string {
	// Replace '==' with '#~~#'
	replacedString := strings.Replace(encodedString, "=", EQUALS_REP, -1)

	// Replace '//' with '()!!'
	replacedString = strings.Replace(replacedString, "/", BACKSLACK_REP, -1)

	return replacedString
}

func restorePaddingChars(replacedString string) string {
	// Restore '#~~#' to '=='
	restoredString := strings.Replace(replacedString, EQUALS_REP, "=", -1)

	// Restore '()!!' to '//'
	restoredString = strings.Replace(restoredString, BACKSLACK_REP, "/", -1)

	return restoredString
}

// EncodeAndCompressURL encodes and compresses the given URL.
func EncodeAndCompressString(str string) (string, error) {

	// Compress the encoded URL
	var compressedURL bytes.Buffer
	compressor := zlib.NewWriter(&compressedURL)
	_, err := compressor.Write([]byte(str))
	if err != nil {
		return "", fmt.Errorf("error compressing URL: %w", err)
	}
	compressor.Close()

	// Convert the compressed URL to ASCII base64 string
	asciiURL := base64.StdEncoding.EncodeToString(compressedURL.Bytes())
	asciiURL = replacePaddingChars(asciiURL)
	return asciiURL, nil
}

// DecompressAndDecodeURL decompresses and decodes the given compressed URL.
func DecompressAndDecodeString(compressedString string) (string, error) {
	// Convert the ASCII base64 string to compressed byte slice
	compressedBytes, err := base64.StdEncoding.DecodeString(compressedString)
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
	res := restorePaddingChars(decompressedBuffer.String())
	return res, nil
}
