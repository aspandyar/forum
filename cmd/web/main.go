package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

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

	err := LoadEnvFromFile(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbUser == "" || dbPassword == "" {
		errorLog.Fatal("Missing database credentials. Set DB_USER and DB_PASSWORD environment variables.")
	}

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

	err = app.createAdmin()
	if err != nil {
		errorLog.Fatal(err)
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
		MaxVersion:       tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	srv := &http.Server{
		Addr:           *addr,
		ErrorLog:       errorLog,
		MaxHeaderBytes: 524288,
		Handler:        app.routes(),
		TLSConfig:      tlsConfig,
		IdleTimeout:    time.Minute,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	infoLog.Printf("starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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

	statements := strings.Split(string(sqlScript), ";")

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
