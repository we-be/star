package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/we-be/star/internal/api"
	"github.com/we-be/star/internal/db"
	"github.com/we-be/star/internal/service"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dsn := flag.String("dsn", "postgres://localhost:5432/star?sslmode=disable", "postgres connection string")
	seed := flag.Uint64("seed", 42, "RNG seed for generation")
	flag.Parse()

	if v := os.Getenv("DATABASE_URL"); v != "" {
		*dsn = v
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	queries := db.New(pool)
	svc := service.New(queries, *seed)

	if err := svc.LoadReferenceData(ctx); err != nil {
		log.Fatalf("load reference data: %v", err)
	}
	fmt.Println("reference data loaded")

	srv := &http.Server{
		Addr:         *addr,
		Handler:      api.New(svc),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second, // generous for universe generation
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nshutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	fmt.Printf("star service listening on %s\n", *addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
