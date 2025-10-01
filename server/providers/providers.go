package providers

import (
	"fmt"
	"os/exec"
)

type ProviderQueryObject struct {
	IMDbID              *string  `json:"imdb_id"`      // starts with 'tt'
	MediaType           string   `json:"media_type"`   // movies or tvshows, etc.
	Season                 int   `json:"season"`
	Episode                int   `json:"episode"`
	Query				*string  `json:"search_query"`
	Params            []string   `json:"params"`
}

type ProviderQueryResponse struct {
	Status              bool    `json:"go_status"`
	IMDbID              string  `json:"imdb_id"`      // starts with 'tt'
	MediaType           string  `json:"media_type"`   // movies or tvshows, etc.

}

func InitializeProviders() {

}

//func SearchProviders(query *ProviderQueryObject) (*map[string]interface{}, error) {
//	if *query.IMDbID != "" {
//
//	}
//	scriptList := []string{"script1.py", "script2.py", "script3.py"}
//	resultChan := make(chan string, len(scriptList))
//
//	// Start a goroutine for each script
//	for _, script := range scriptList {
//		go runScript(script, resultChan)
//	}
//	channel := ""
//	// Listen for results on the channel and stream them to the client
//	for i := 0; i < len(scriptList); i++ {
//		channel = <-resultChan
//	}
//
//	//cmd := exec.Command("python", "providers/torrentio.py")
//	//output, err := cmd.CombinedOutput()
//	//if err != nil {
//	//	fmt.Println("Error running script:", err)
//	//	return nil, err
//	//}
//	//var result map[string]interface{}
//	//if err := json.Unmarshal(result, &result); err != nil {
//	//	fmt.Println("Error parsing JSON:", err)
//	//	return nil, err
//	//}
//	return &result, nil
//}

func runScript(filepath string, resultChan chan<- string) {
	cmd := exec.Command("python", filepath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		resultChan <- fmt.Sprintf("Error running %s: %v", filepath, err)
		return
	}
	resultChan <- fmt.Sprintf("Output from %s: %s", filepath, string(output))
}