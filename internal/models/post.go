package models

import "time"

type Post struct {
	ID         int
	UserID     int
	Title      string
	Content    string
	Categories []Category
	CreatedAt  time.Time
	Author     string
	Likes      int
	Dislikes   int
}

type Category struct {
	ID   int
	Name string
}
