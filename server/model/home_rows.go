package model

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"
	"github.com/mcay23/hound/view"
)

const (
	MaxItemsPerHomeRow = 20
	HomeRowRefreshTime = time.Hour * 24 // refresh home row every 24 hours
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
	switch homeRow.CatalogSelection {
	case database.CatalogSelectionRotate:
		// rotate to a new catalog index every day
		days := int(time.Now().Unix() / 86400)
		selectedCatalog := homeRow.Catalogs[days%len(homeRow.Catalogs)]
		catalogItems, err = GetCatalog(selectedCatalog.CatalogSource, selectedCatalog.CatalogID, userID)
		if err != nil {
			return nil, err
		}
		if selectedCatalog.CatalogTitle != "" {
			title = selectedCatalog.CatalogTitle
		}
	case database.CatalogSelectionAll:
		for _, catalog := range homeRow.Catalogs {
			temp, err := GetCatalog(catalog.CatalogSource, catalog.CatalogID, userID)
			if err != nil {
				return nil, err
			}
			catalogItems = append(catalogItems, temp...)
		}
	default:
		return nil, fmt.Errorf("invalid home row selection type: %s: %w", homeRow.CatalogSelection, internal.BadRequestError)
	}
	catalogItems = deduplicateCatalogItems(userID, rowIndex, catalogItems, homeRow.ItemOrder == database.ItemOrderRandom)
	return &view.HomeRowView{
		Title: title,
		Items: catalogItems,
	}, nil
}

func deduplicateCatalogItems(userID int64, rowIndex int, viewArray []view.MediaRecordCatalog, randomize bool) []view.MediaRecordCatalog {
	if len(viewArray) == 0 {
		return []view.MediaRecordCatalog{}
	}
	// Build deterministic seed, so internal catalog randomization is stable if there are no
	// additional changes, since internal catalogs are not cached (eg. Hound Library/Downloads)
	// this will break if rowIndex changes, but not really that critical
	day := time.Now().UTC().Unix() / 86400
	hash := fnv.New64a()
	hash.Write([]byte(strconv.FormatInt(userID, 10)))
	hash.Write([]byte(":"))
	hash.Write([]byte(strconv.Itoa(rowIndex)))
	hash.Write([]byte(":"))
	hash.Write([]byte(strconv.FormatInt(day, 10)))
	seed := int64(hash.Sum64())
	r := rand.New(rand.NewSource(seed))

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
		r.Shuffle(len(deduped), func(i, j int) {
			deduped[i], deduped[j] = deduped[j], deduped[i]
		})
	}
	limit := MaxItemsPerHomeRow
	if limit > len(deduped) {
		limit = len(deduped)
	}
	return deduped[:limit]
}
