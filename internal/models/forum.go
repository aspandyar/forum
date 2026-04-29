package models

import (
	"database/sql"
	"io/ioutil"
	"log"
	"time"
)

type Forum struct {
	ID            int
	Title         string
	Content       string
	Tags          string
	TagsOutput    []string
	Reacted       bool
	Liked         bool
	LikesCount    int
	DislikesCount int
	Comment       []userComment
	Created       time.Time
	Expires       time.Time
	ImagePath     string
	IsOwnForum    bool
	EditComment   userComment
}

type userComment struct {
	CommentID     int
	ForumID       int
	User          string
	Comment       string
	Reacted       bool
	Liked         bool
	LikesCount    int
	DislikesCount int
	IsOwnComment  bool
}

type ForumModel struct {
	DB *sql.DB
}

func (m *ForumModel) Init(initSqlFileName string) error {
	file, err := ioutil.ReadFile(initSqlFileName)
	if err != nil {
		log.Fatalf("Can't read SQL file %v", err)
	}

	_, err = m.DB.Exec(string(file))
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	return nil
}
