package models


type Submition struct {
	Id string
	Gif bool
	Author string
	Timestamp int64
	Content string
}

type User struct {
	Username string
	Password string
	MaxMb int8
	Submitted []string
	Id string
	Admin bool
}