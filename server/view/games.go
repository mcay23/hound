package view

import "github.com/mcay23/hound/sources"

type GameFullObject struct {
	*sources.IGDBGameObject
	Comments *[]CommentObject `json:"comments"`
}
