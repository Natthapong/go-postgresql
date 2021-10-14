package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
)

type Book struct {
	Isbn   string  `json:"isbn"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float32 `json:"price"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:passw0rd@localhost/demo?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println("go-postgresql")
	// Setup
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.GET("/books", booksIndex)

	// Start server
	go func() {
		if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func getAllBook() ([]Book, error) {
	var books []Book
	rows, err := db.Query(`SELECT * FROM books`)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Book
		err = rows.Scan(&b.Isbn, &b.Title, &b.Author, &b.Price)
		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}
		books = append(books, b)

	}
	return books, err
}

func booksIndex(c echo.Context) error {
	books, err := getAllBook()
	if err != nil {
		log.Fatal(err)
	}
	return c.JSON(http.StatusOK, books)
}
