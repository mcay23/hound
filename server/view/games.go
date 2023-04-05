package view

import "hound/model/sources"

type GameFullObject struct {
	*sources.IGDBGameObject
	Comments         *[]CommentObject         `json:"comments"`
}
