package sources

import "fmt"

func InitializeSources() {
	InitializeTMDB()
	test, err := GetGameFromIDIGDB(121)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(test.Name)
}
