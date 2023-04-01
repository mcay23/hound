package sources

import (
	"fmt"
	"github.com/Henry-Sarabia/igdb/v2"
	"os"
)

const (
	SourceIGDB string = "igdb"
)

var igdbClient *igdb.Client

func InitializeIGDB() {
	fmt.Println(os.Getenv("IGDB_CLIENT_ID"), os.Getenv("IGDB_CLIENT_SECRET"))
	igdbClient = igdb.NewClient(os.Getenv("IGDB_CLIENT_ID"), os.Getenv("IGDB_CLIENT_SECRET"), nil)
	tmdbClient.SetClientAutoRetry()
	games, err := igdbClient.Games.Search("zelda")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, item := range games {
			fmt.Println(item.Name)
		}
		fmt.Println(games)
	}
}