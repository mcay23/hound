package helpers

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*]`)

func SanitizeFilename(filename string) string {
	return invalidFilenameChars.ReplaceAllString(filename, "")
}

func PrettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(InfoMsg(string(s)))
}
