package model

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/view"
)

const (
	MaxItemsPerHomeRow = 20
)

/*
Evaluates and returns each home row defined by rowIndex
eg. User has home rows ["trending-movies","trending-shows"]

Clients should request the user's home rows,
and make 2 HTTP requests:

	-> rowIndex 0 (trending movies)
	-> rowIndex 1 (trending shows)
*/
func GetHomeRow(userID int64, rowIndex int) (*view.HomeRowView, error) {
	cacheKey := fmt.Sprintf("user_home_row|%d|%d", userID, rowIndex)
	var homeRowView view.HomeRowView
	found, _ := database.GetCache(cacheKey, &homeRowView)
	if found {
		return &homeRowView, nil
	}
	userHomeRows, err := database.GetUserHomeRows(userID)
	if err != nil {
		return nil, err
	}
	if rowIndex < 0 || rowIndex >= len(userHomeRows.HomeRows) {
		return nil, fmt.Errorf("invalid home row index: %d: %w", rowIndex, internal.BadRequestError)
	}
	homeRow := userHomeRows.HomeRows[rowIndex]
	catalogItems := []view.MediaRecordCatalog{}
	title := homeRow.Title
	switch homeRow.SelectionType {
	case database.SelectionTypeRotate:
		// rotate to a new catalog index every day
		days := int(time.Now().Unix() / 86400)
		selectedCatalog := homeRow.Catalogs[days%len(homeRow.Catalogs)]
		catalogs, err := GetCatalog(selectedCatalog.CatalogSource, selectedCatalog.CatalogID, userID)
		if err != nil {
			return nil, err
		}
		if selectedCatalog.CatalogTitle != "" {
			title = selectedCatalog.CatalogTitle
		}
		catalogItems = deduplicateCatalogItems(catalogs, false)
	case database.SelectionTypeMix:
		var viewArray []view.MediaRecordCatalog
		for _, catalog := range homeRow.Catalogs {
			catalogViewArray, err := GetCatalog(catalog.CatalogSource, catalog.CatalogID, userID)
			if err != nil {
				return nil, err
			}
			viewArray = append(viewArray, catalogViewArray...)
		}
		// select up to MaxItemsPerHomeRow unique items randomly
		catalogItems = deduplicateCatalogItems(viewArray, true)
	default:
		return nil, fmt.Errorf("invalid home row selection type: %s: %w", homeRow.SelectionType, internal.BadRequestError)
	}
	if len(catalogItems) > 0 {
		_, _ = database.SetCache(cacheKey, view.HomeRowView{
			Title: title,
			Items: catalogItems,
		}, database.CatalogRotationTime)
	}
	return &view.HomeRowView{
		Title: title,
		Items: catalogItems,
	}, nil
}

func deduplicateCatalogItems(viewArray []view.MediaRecordCatalog, randomize bool) []view.MediaRecordCatalog {
	if len(viewArray) == 0 {
		return []view.MediaRecordCatalog{}
	}
	// Deduplicate by ID
	seen := make(map[string]struct{})
	deduped := make([]view.MediaRecordCatalog, 0, len(viewArray))
	for _, item := range viewArray {
		key := item.MediaSource + "-" + item.MediaType + "-" + item.SourceID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, item)
	}
	if randomize {
		rand.Shuffle(len(deduped), func(i, j int) {
			deduped[i], deduped[j] = deduped[j], deduped[i]
		})
	}
	limit := MaxItemsPerHomeRow
	if limit > len(deduped) {
		limit = len(deduped)
	}
	return deduped[:limit]
}
