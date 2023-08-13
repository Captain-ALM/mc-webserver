package utils

import (
	"crypto"
	"encoding/hex"
	"strings"
)

func GetValueForETagUsingByteArray(b []byte) string {
	theHash := crypto.SHA1.New()
	_, _ = theHash.Write(b)
	theSum := theHash.Sum(nil)
	theHash.Reset()
	return "\"" + hex.EncodeToString(theSum) + "\""
}

func GetETagValues(stringIn string) []string {
	if strings.ContainsAny(stringIn, ",") {
		seperated := strings.Split(stringIn, ",")
		toReturn := make([]string, len(seperated))
		pos := 0
		for _, s := range seperated {
			cETag := GetETagValue(s)
			if cETag != "" {
				toReturn[pos] = cETag
				pos += 1
			}
		}
		if pos == 0 {
			return nil
		}
		return toReturn[:pos]
	}
	toReturn := []string{GetETagValue(stringIn)}
	if toReturn[0] == "" {
		return nil
	}
	return toReturn
}

func GetETagValue(stringIn string) string {
	startIndex := strings.IndexAny(stringIn, "\"") + 1
	endIndex := strings.LastIndexAny(stringIn, "\"")
	if endIndex > startIndex {
		return stringIn[startIndex:endIndex]
	}
	return ""
}
