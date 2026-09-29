package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rishikdurvasula/aegis/internal/persistence"
)

func main() {
	ctx := context.Background()

	databaseURL := "postgres://localhost/aegis?sslmode=disable"

	db, err := persistence.NewDB(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Connected to PostgreSQL!")
}
