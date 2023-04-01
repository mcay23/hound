package helpers

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(InfoMsg(string(s)))
}
