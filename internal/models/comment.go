package models

import "time"

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Author    string
	Content   string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
}
