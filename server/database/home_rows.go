package database

import (
	"fmt"
	"time"
)

type UserHomeRows struct {
	UserID      int64     `json:"user_id"`
	LastUpdated time.Time `json:"last_updated"`
	HomeRows    []HomeRow `json:"home_rows" binding:"required,min=1,dive"`
}

/*
Represents a row in the home screen. This is distinct from a catalog, in that
a single row can be composed of multiple catalogs.

eg. You mix two catalogs - crime shows and crime movies into a single row

Row Order - "random" - randomizes order of catalogs || "rotate" - rotate between catalogs, only one is shown at a time
  - "default" - follows order of catalogs in array

Title - The default title, used if row_order is mix
*/
type HomeRow struct {
	Title            string    `json:"title" binding:"required,gt=0"`
	CatalogSelection string    `json:"catalog_selection" binding:"required,oneof=all rotate mix"`
	ItemOrder        string    `json:"item_order" binding:"required,oneof=default random"`
	Catalogs         []Catalog `json:"catalogs" binding:"required,min=1,dive"`
}

const (
	CatalogSourceInternal  = "internal"
	CatalogSourceTMDB      = "tmdb"
	CatalogSourceMDBList   = "mdblist"
	CatalogSelectionRotate = "rotate"  // one catalog is shown at a time, rotated every period
	CatalogSelectionAll    = "all"     // all catalogs are shown
	ItemOrderDefault       = "default" // follows order of catalog
	ItemOrderRandom        = "random"  // randomize order of catalog
)

type Catalog struct {
	CatalogTitle  string `json:"catalog_title"` // Prioritized if selection_type is rotate
	CatalogSource string `json:"catalog_source" binding:"required,oneof=internal tmdb mdblist"`
	CatalogID     string `json:"catalog_id" binding:"required,gt=0"`
}

func GetUserHomeRows(userID int64) (*UserHomeRows, error) {
	var userHomeRows UserHomeRows
	found, _ := GetCache(fmt.Sprintf("user_home_rows|%d", userID), &userHomeRows)
	if found {
		return &userHomeRows, nil
	}
	defaultHomeRows, err := GetDefaultHomeRows()
	if err != nil {
		return nil, err
	}
	userHomeRows.HomeRows = defaultHomeRows.HomeRows
	userHomeRows.UserID = userID
	return &userHomeRows, nil
}

// Default home rows for all users
func GetDefaultHomeRows() (*UserHomeRows, error) {
	var defaultHomeRow UserHomeRows
	found, _ := GetCache("default_home_rows", &defaultHomeRow)
	if found {
		return &defaultHomeRow, nil
	}
	defaultHomeRow.HomeRows = []HomeRow{
		{
			Title:            "Trending Shows",
			CatalogSelection: CatalogSelectionAll,
			ItemOrder:        ItemOrderDefault,
			Catalogs: []Catalog{
				{
					CatalogTitle:  "Trending Shows",
					CatalogSource: CatalogSourceTMDB,
					CatalogID:     "trending-shows",
				},
			},
		},
		{
			Title:            "Trending Movies",
			CatalogSelection: CatalogSelectionAll,
			ItemOrder:        ItemOrderDefault,
			Catalogs: []Catalog{
				{
					CatalogTitle:  "Trending Movies",
					CatalogSource: CatalogSourceTMDB,
					CatalogID:     "trending-movies",
				},
			},
		},
		{
			Title:            "Movies You Might Like",
			CatalogSelection: CatalogSelectionAll,
			ItemOrder:        ItemOrderDefault,
			Catalogs: []Catalog{
				{
					CatalogTitle:  "Movies You Might Like",
					CatalogSource: CatalogSourceInternal,
					CatalogID:     "hound-recommended-movies",
				},
			},
		},
		{
			Title:            "Recently Added to Hound",
			CatalogSelection: CatalogSelectionAll,
			ItemOrder:        ItemOrderDefault,
			Catalogs: []Catalog{
				{
					CatalogTitle:  "Recently Added to Hound",
					CatalogSource: CatalogSourceInternal,
					CatalogID:     "hound-library",
				},
			},
		},
		{
			Title:            "From Netflix",
			CatalogSelection: CatalogSelectionAll,
			ItemOrder:        ItemOrderRandom,
			Catalogs: []Catalog{
				{
					CatalogTitle:  "From Netflix",
					CatalogSource: CatalogSourceTMDB,
					CatalogID:     "netflix-movies",
				},
				{
					CatalogTitle:  "From Netflix",
					CatalogSource: CatalogSourceTMDB,
					CatalogID:     "netflix-shows",
				},
			},
		},
		{
			Title:            "Animation Movies",
			CatalogSelection: CatalogSelectionAll,
			ItemOrder:        ItemOrderRandom,
			Catalogs: []Catalog{
				{
					CatalogTitle:  "Animation Movies",
					CatalogSource: CatalogSourceTMDB,
					CatalogID:     "genre-movies-19",
				},
			},
		},
	}
	defaultHomeRow.UserID = -1
	defaultHomeRow.LastUpdated = time.Now()
	_, err := SetCache("default_home_rows", defaultHomeRow, -1)
	if err != nil {
		return nil, err
	}
	return &defaultHomeRow, nil
}

func UpdateUserHomeRows(homeRow UserHomeRows) error {
	homeRow.LastUpdated = time.Now()
	_, err := SetCache(fmt.Sprintf("user_home_rows|%d", homeRow.UserID), homeRow, -1)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDefaultHomeRows(homeRow UserHomeRows) error {
	homeRow.UserID = -1
	homeRow.LastUpdated = time.Now()
	_, err := SetCache("default_home_rows", homeRow, -1)
	if err != nil {
		return err
	}
	return nil
}
