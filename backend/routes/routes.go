package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kushalpraja/library-api/handlers"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/books/list", handlers.GetBooks)
	r.POST("/books/add", handlers.AddBook)
	r.PATCH("/books/edit", handlers.EditBook)
	r.DELETE("/books/delete", handlers.DeleteBook)
}
