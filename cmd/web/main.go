package main

import (
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/aspandyar/forum/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	forums        *models.ForumModel
	sessions      *models.SessionModel
	users         *models.UserModel
	forumLike     *models.ForumLikesModel
	forumComment  *models.ForumCommentModel
	tempalteCache map[string]*template.Template
}

const (
	newDbName       = "./st.db"
	initSqlFileName = "./init-up.sql"
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(newDbName)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	readDB(initSqlFileName, db)

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		forums:        &models.ForumModel{DB: db},
		users:         &models.UserModel{DB: db},
		sessions:      &models.SessionModel{DB: db},
		forumLike:     &models.ForumLikesModel{DB: db},
		forumComment:  &models.ForumCommentModel{DB: db},
		tempalteCache: templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func readDB(path string, db *sql.DB) {
	sqlScript, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	// Split the script into individual statements
	statements := strings.Split(string(sqlScript), ";")

	// Execute each SQL statement
	for _, stmt := range statements {
		trimmedStmt := strings.TrimSpace(stmt)
		if len(trimmedStmt) > 0 {
			_, err := db.Exec(trimmedStmt)
			if err != nil {
				log.Println("Error executing statement:", err)
			}
		}
	}
}
