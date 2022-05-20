package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Book struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	dataSourceName := fmt.Sprintf("host=db port=5432 user=postgres password=%s dbname=postgres sslmode=disable", os.Getenv("POSTGRES_PASSWORD"))
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		fmt.Println(err)
	}

	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM books")

		if err != nil {
			fmt.Println(err)
			return
		}

		var books []Book
		for rows.Next() {
			var book Book
			err := rows.Scan(&book.ID, &book.Name)
			if err != nil {
				fmt.Println(err)
				return
			}

			books = append(books, book)

		}

		render.JSON(w, r, books)
	})

	http.ListenAndServe(":8080", r)
}
