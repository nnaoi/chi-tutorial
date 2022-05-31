package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	models "chi-tutorial/models"
)

type Book struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type key string

var (
	db      *sql.DB
	once    sync.Once
	bookKey key = "book"
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

	r.Route("/books", func(r chi.Router) {
		r.Get("/", ListBooks)
		r.Post("/", CreateBook)

		r.Route("/{bookID}", func(r chi.Router) {
			r.Use(BookCtx)
			r.Get("/", GetBook)
			r.Put("/", UpdateBook)
			r.Delete("/", DeleteBook)
		})
	})

	http.ListenAndServe(":8080", r)
}

func ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := models.Books(qm.OrderBy("id")).All(context.Background(), db)
	if err != nil {
		fmt.Println(err)
		return
	}

	render.JSON(w, r, books)
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	book := &models.Book{}
	json.NewDecoder(r.Body).Decode(&book)
	err := book.Insert(context.Background(), db, boil.Infer())
	if err != nil {
		fmt.Println(err)
		return
	}

	render.JSON(w, r, book)
}

// BookCtx middleware is used to load an book object from
// the URL parameters passed through as the request. In case
// the Book could not be found, we stop here and return a 404.
func BookCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var err error
		bookID := chi.URLParam(r, "bookID")
		if bookID == "" {
			fmt.Println("bookID empty")
			return
		}

		intBookID, err := strconv.Atoi(bookID)
		if err != nil {
			fmt.Println(err)
			return
		}

		book, err := models.FindBook(context.Background(), db, intBookID)
		if err != nil {
			fmt.Println(err)
			return
		}

		ctx := context.WithValue(r.Context(), bookKey, book)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetBook returns the specific Book. You'll notice it just
// fetches the Book right off the context, as its understood that
// if we made it this far, the Book must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func GetBook(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the article
	// context because this handler is a child of the BookCtx
	// middleware. The worst case, the recoverer middleware will save us.
	book := r.Context().Value(bookKey).(*models.Book)
	render.JSON(w, r, book)
}

// UpdateBook updates an existing Book in our persistent store.
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	book := r.Context().Value(bookKey).(*models.Book)
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = book.Update(context.Background(), db, boil.Infer())
	if err != nil {
		fmt.Println(err)
		return
	}

	render.JSON(w, r, book)
}

// DeleteBook removes an existing Book from our persistent store.
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	book := r.Context().Value(bookKey).(*models.Book)
	_, err := book.Delete(context.Background(), db)
	if err != nil {
		fmt.Println(err)
		return
	}

	books, err := models.Books().All(context.Background(), db)
	if err != nil {
		fmt.Println(err)
		return
	}

	render.JSON(w, r, books)
}
