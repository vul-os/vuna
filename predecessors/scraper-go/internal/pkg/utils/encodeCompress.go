package utils

import (
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
	productId := RemoveWWWPrefix(RemoveHTTPPrefix(rawURL))
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

func StringToIdentifier(urlString string, otherStringIds []string) (string, string, error) {
	hostString, productString, err := GetHostName(urlString)

	if err != nil {
		return "", "", err
	}
	hostId := StringToSiteIdentifier(hostString)

	productId, err := StringToProductIdentifier(productString, otherStringIds)
	if err != nil {
		return "", "", err
	}
	return hostId, productId, nil
}

func StringToProductIdentifier(stringy string, otherStringIds []string) (string, error) {
	fmt.Println("here: ", stringy, otherStringIds)
	stringsToJoin := append(otherStringIds, stringy)
	joinedString := strings.Join(stringsToJoin, CHAR_COMBO)
	encoded4URL := url.QueryEscape(joinedString)
	fmt.Println("here2: ", encoded4URL)
	return EncodeString(encoded4URL), nil
}

func StringToSiteIdentifier(stringy string) string {
	return EncodeString(stringy)
}

func EncodeURL(urlStr string) string {
	encodedURL := url.QueryEscape(urlStr)
	encodedURL = base64.URLEncoding.EncodeToString([]byte(encodedURL))
	return encodedURL
}

func EncodeString(str string) string {
	encodedStr := base64.StdEncoding.EncodeToString([]byte(str))
	return replacePaddingChars(encodedStr)
}

func replacePaddingChars(encodedString string) string {
	replacedString := strings.Replace(encodedString, "=", EQUALS_REP, -1)
	replacedString = strings.Replace(replacedString, "/", BACKSLACK_REP, -1)
	return replacedString
}

func restorePaddingChars(replacedString string) string {
	restoredString := strings.Replace(replacedString, EQUALS_REP, "=", -1)
	restoredString = strings.Replace(restoredString, BACKSLACK_REP, "/", -1)
	return restoredString
}
