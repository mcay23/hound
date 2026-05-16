package database

import (
	"fmt"
	"time"
)

type UserHomeRows struct {
	UserID      int64     `json:"user_id"`
	LastUpdated time.Time `json:"last_updated"`
	HomeRows    []HomeRow `json:"home_row"`
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
	Title            string    `json:"title"`
	CatalogSelection string    `json:"catalog_selection"`
	ItemOrder        string    `json:"item_order"`
	Catalogs         []Catalog `json:"catalogs"`
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
	CatalogSource string `json:"catalog_source"`
	CatalogID     string `json:"catalog_id"`
}

func GetUserHomeRows(userID int64) (*UserHomeRows, error) {
	var userHomeRows UserHomeRows
	found, _ := GetCache(fmt.Sprintf("user_home_rows|%d", userID), &userHomeRows)
	if found {
		return &userHomeRows, nil
	}
	return GetDefaultHomeRows(userID)
}

// Default home rows for all users
func GetDefaultHomeRows(userID int64) (*UserHomeRows, error) {
	var userHomeRows UserHomeRows
	// found, _ := GetCache("default_home_rows", &userHomeRows)
	// if found {
	// 	userHomeRows.UserID = userID
	// 	return &userHomeRows, nil
	// }
	userHomeRows.HomeRows = []HomeRow{
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
			Title:            "Recently Added to Hound",
			CatalogSelection: CatalogSelectionAll,
			ItemOrder:        ItemOrderDefault,
			Catalogs: []Catalog{
				{
					CatalogTitle:  "Recently Added to Hound",
					CatalogSource: CatalogSourceInternal,
					CatalogID:     "hound-library-movies",
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
					CatalogID:     "genre-movies-2",
				},
			},
		},
	}
	userHomeRows.LastUpdated = time.Now()
	userHomeRows.UserID = userID
	_, err := SetCache("default_home_rows", userHomeRows, -1)
	if err != nil {
		return nil, err
	}
	return &userHomeRows, nil
}
