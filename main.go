package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"minder/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True",
		os.Getenv("MYSQL_USERNAME"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("cannot establish database connection: %s", dsn)
		return
	}

	if err = auth.SetDB(db); err != nil {
		log.Fatalf("cannot set dependency for auth: %s", err.Error())
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/auth", auth.Router())

	srvPort := os.Getenv("SERVER_PORT")

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", srvPort),
		Handler: r,
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func(s *http.Server) {
		log.Printf("goodies http now available at %s\n", s.Addr)
		if serr := s.ListenAndServe(); serr != http.ErrServerClosed {
			log.Fatal(serr)
		}
	}(s)

	<-sigChan

	err = s.Shutdown(context.Background())
	if err != nil {
		log.Fatal("something wrong when stopping server : ", err)
		return
	}

	log.Printf("server stopped")
}
