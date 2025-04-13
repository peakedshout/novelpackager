package bilinovel

import (
	"slices"
)

func decryptionFont(str string) string {
	var data string
	for _, r := range str {
		before := string(r)
		if slices.Contains(blankUnicode, before) {
			continue
		}
		f, ok := fontSecretMap[before]
		if ok {
			data += f
		} else {
			data += before
		}
	}
	return data
}
