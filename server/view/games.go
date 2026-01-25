package view

import "hound/sources"

type GameFullObject struct {
	*sources.IGDBGameObject
	Comments *[]CommentObject `json:"comments"`
}
