package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"forum/config"
	"forum/internal/models"
	"forum/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	posts          *models.PostModel
	templateCache  map[string]*template.Template
	users          *models.UserModel
	sessionManager *SessionManager
	postTags       *models.CategoryModel
	comments       *models.CommentModel
	reactions      *models.ReactionModel
}

func main() {
	config, err := config.NewConfig()
	if err != nil {
		return
	}
	db, err := store.NewSqlite3(config)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	defer db.Close()
	f, err := os.OpenFile("./info.log", os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		posts:          &models.PostModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		sessionManager: &SessionManager{DB: db},
		postTags:       &models.CategoryModel{DB: db},
		comments:       &models.CommentModel{DB: db},
		reactions:      &models.ReactionModel{DB: db},
	}
	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:     ":8081",
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", server.Addr)
	err = server.ListenAndServe()
	errorLog.Fatal(err)
}
