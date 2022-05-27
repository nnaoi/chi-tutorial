package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Book struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

var (
	db   *sql.DB
	once sync.Once
)

// getDB lazily instantiates a database connection pool. Users of Cloud Run or
// Cloud Functions may wish to skip this lazy instantiation and connect as soon
// as the function is loaded. This is primarily to help testing.
func getDB() *sql.DB {
	once.Do(func() {
		db = mustConnect()
	})
	return db
}

// mustConnect creates a connection to the database based on environment
// variables. Setting the optional DB_CONN_TYPE environment variable to UNIX or
// TCP will use the corresponding connection method. By default, the connector
// is used.
func mustConnect() *sql.DB {
	var (
		db  *sql.DB
		err error
	)

	onCloudRun := os.Getenv("INSTANCE_UNIX_SOCKET") != ""
	// Use a Unix socket when INSTANCE_UNIX_SOCKET (e.g., /cloudsql/proj:region:instance) is defined.
	if onCloudRun {
		db, err = connectUnixSocket()
		if err != nil {
			log.Fatalf("connectUnixSocket: unable to connect: %s", err)
		}
	}

	if db == nil {
		// Connect to localhost DB.
		db, err = connectToLocal()
	}
	if err != nil {
		log.Fatalf("localhost: unable to connect: %s", err)
	}

	return db
}

func main() {
	db := getDB()
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

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
