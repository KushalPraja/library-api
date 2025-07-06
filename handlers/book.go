package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/kushalpraja/library-api/db"
	"github.com/kushalpraja/library-api/models"
	"net/http"
	"strconv"
)

func GetBooks(c *gin.Context) {
	// select all books from the library table
	rows, err := db.DB.Query("SELECT Book_name, Author, ISBN FROM library")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.BookName, &book.Author, &book.ISBN); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}
	c.IndentedJSON(http.StatusOK, books)
}

func AddBook(c *gin.Context) {
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("INSERT INTO library (Book_name, Author, ISBN) VALUES (?, ?, ?)",
		book.BookName, book.Author, book.ISBN)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Book added successfully"})
}

func EditBook(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
		Field string `json:"field"`
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query string
	var value any

	switch req.Field {
	case "Book_name", "Author":
		query = "UPDATE library SET " + req.Field + " = ? WHERE Book_name = ?"
		value = req.Value
	case "ISBN":
		intVal, err := strconv.Atoi(req.Value)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid ISBN"})
			return
		}
		query = "UPDATE library SET ISBN = ? WHERE Book_name = ?"
		value = intVal
	default:
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid field"})
		return
	}

	result, err := db.DB.Exec(query, value, req.Title)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Book updated"})
}

func DeleteBook(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.DB.Exec("DELETE FROM library WHERE Book_name = ?", req.Title)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Book deleted"})
}
