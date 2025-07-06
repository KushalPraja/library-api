package models

type Book struct {
	BookName string `json:"book_name"`
	Author   string `json:"author"`
	ISBN     int    `json:"isbn"`
}
