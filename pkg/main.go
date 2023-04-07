package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	API_PATH = "/api/v1/books"
)

type library struct {
	dbHost, dbPass, dbName string
}

type Book struct {
	Id, Name, Isbn string
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost:3306"
	}

	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "Sh@260103"
	}

	api_path := os.Getenv("API_PATH")
	if api_path == "" {
		api_path = API_PATH
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "library"
	}

	l := library{
		dbHost: dbHost,
		dbPass: dbPass,
		dbName: dbName,
	}
	r := mux.NewRouter()
	r.HandleFunc(API_PATH, l.getBooks).Methods("GET")
	r.HandleFunc(API_PATH, l.postBook).Methods("POST")
	http.ListenAndServe(":8080", r)
}

func (l library) getBooks(w http.ResponseWriter, r *http.Request) {
	db := l.openConnection()
	rows, err := db.Query("select * from books")
	if err != nil {
		log.Fatal(err.Error())
	}

	books := []Book{}
	for rows.Next() {
		var id, name, isbn string
		err := rows.Scan(&id, &name, &isbn)
		if err != nil {
			log.Fatal("Error occured while scanning row", err.Error())
		}
		oBook := Book{
			Id:   id,
			Name: name,
			Isbn: isbn,
		}

		books = append(books, oBook)
	}

	json.NewEncoder(w).Encode(books)
	l.closeConnection(db)
}

func (l library) postBook(w http.ResponseWriter, r *http.Request) {
	db := l.openConnection()
	var book Book
	json.NewDecoder(r.Body).Decode(&book)
	insertQuery, err := db.Prepare("insert into books values(?,?,?)")
	if err != nil {
		log.Fatal("Preparing the database query")
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("While begining the transaction", err.Error())
	}

	_, err = tx.Stmt(insertQuery).Exec(book.Id, book.Name, book.Isbn)

	if err != nil {
		log.Fatal("excuting the insert command ", err.Error())
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal("While Commiting the transaction ", err.Error())
	}
	json.NewEncoder(w).Encode("Book added")
	l.closeConnection(db)
}
func (l library) openConnection() *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s", "root", l.dbPass, l.dbHost, l.dbName))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func (l library) closeConnection(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatal("Connection was unable to close")
	}
}
