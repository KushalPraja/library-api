package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
)

type library struct {
	Book_name string `json:"book_name"`
	Author    string `json:"author"`
	ISBN      int    `json:"isbn"`
}

var db *sql.DB

func getBook(x *gin.Context) {
	rows, err := db.Query("SELECT Book_name, Author, ISBN FROM library")
	if err != nil {
		x.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []library
	for rows.Next() {
		var book library
		err := rows.Scan(&book.Book_name, &book.Author, &book.ISBN)
		if err != nil {
			x.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}
	x.IndentedJSON(http.StatusOK, books)
}

func addBook(x *gin.Context) {
	var book library
	if err := x.ShouldBindJSON(&book); err != nil {
		x.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("INSERT INTO library (Book_name, Author, ISBN) VALUES (?, ?, ?)",
		book.Book_name, book.Author, book.ISBN)
	if err != nil {
		x.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	x.IndentedJSON(http.StatusOK, gin.H{"status": "success"})
}

func editBook(x *gin.Context) {

	var req struct {
		Title string `json:"title"`
		Field string `json:"field"`
		Value string `json:"value"`
	}

	if err := x.ShouldBindJSON(&req); err != nil {
		x.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query string
	var value any

	switch req.Field {
	case "Book_name":
		query = "UPDATE library SET Book_name = ? WHERE Book_name = ?"
		value = req.Value
	case "Author":
		query = "UPDATE library SET Author = ? WHERE Book_name = ?"
		value = req.Value
	case "ISBN":
		intValue, err := strconv.Atoi(req.Value)
		if err != nil {
			x.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid ISBN value"})
			return
		}
		query = "UPDATE library SET ISBN = ? WHERE Book_name = ?"
		value = intValue
	default:
		x.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid field"})
		return
	}

	result, err := db.Exec(query, value, req.Title)

	if err != nil {
		x.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		x.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		x.IndentedJSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	x.IndentedJSON(http.StatusOK, gin.H{"status": "success"})
}

type deleteRequest struct {
	Title string `json:"title"`
}

func deleteBook(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("DELETE FROM library WHERE Book_name = ?", req.Title)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"status": "success"})
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to the sqlite database")

	createTableSQL := `CREATE TABLE IF NOT EXISTS library (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		Book_name TEXT NOT NULL,
		Author TEXT NOT NULL,
		ISBN INTEGER
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/Books/list", getBook)
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	router.POST("/Books/add", addBook)
	router.DELETE("/Books/delete", deleteBook)
	router.PATCH("/Books/edit", editBook)
	router.Run(":8080")
}
