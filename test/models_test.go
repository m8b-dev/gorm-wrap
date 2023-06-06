package main

import "gorm.io/gorm"

type Author struct {
	gorm.Model

	Username string
	Posts    []Post
}

type Post struct {
	gorm.Model

	Title   string
	Content string
	Images  []Img
	Videos  []Vid

	AuthorId uint
}

type Img struct {
	gorm.Model

	Title string

	PostId uint
}

type Vid struct {
	gorm.Model

	Title string

	PostId uint
}
