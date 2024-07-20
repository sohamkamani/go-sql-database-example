// file: main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	// Importing pgx v5 for PostgreSQL database operations. The pgx package is used
	// directly for database connection and operations, replacing the standard database/sql package.
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Bird struct {
	Species     string
	Description string
}

func main() {
	// The `sql.Open` function opens a new `*sql.DB` instance. We specify the driver name
	// and the URI for our database. Here, we're using a Postgres URI from an environment variable
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	// Maximum Idle Connections
	db.SetMaxIdleConns(5)
	// Maximum Open Connections
	db.SetMaxOpenConns(10)
	// Idle Connection Timeout
	db.SetConnMaxIdleTime(1 * time.Second)
	// Connection Lifetime
	db.SetConnMaxLifetime(30 * time.Second)

	// To verify the connection to our database instance, we can call the `Ping`
	// method with a context. If no error is returned, we can assume a successful connection
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Println("database is reachable")

	queryCancellation(db)
}

func queryRow(db *sql.DB) {
	// `QueryRow` always returns a single row from the database
	row := db.QueryRow("SELECT bird, description FROM birds LIMIT 1")
	// Create a new `Bird` instance to hold our query results
	bird := Bird{}
	// the retrieved columns in our row are written to the provided addresses
	// the arguments should be in the same order as the columns defined in
	// our query
	if err := row.Scan(&bird.Species, &bird.Description); err != nil {
		log.Fatalf("could not scan row: %v", err)
	}
	fmt.Printf("found bird: %+v\n", bird)
}

func queryRows(db *sql.DB) {
	rows, err := db.Query("SELECT bird, description FROM birds limit 10")
	if err != nil {
		log.Fatalf("could not execute query: %v", err)
	}
	// create a slice of birds to hold our results
	birds := []Bird{}

	// iterate over the returned rows
	// we can go over to the next row by calling the `Next` method, which will
	// return `false` if there are no more rows
	for rows.Next() {
		bird := Bird{}
		// create an instance of `Bird` and write the result of the current row into it
		if err := rows.Scan(&bird.Species, &bird.Description); err != nil {
			log.Fatalf("could not scan row: %v", err)
		}
		// append the current instance to the slice of birds
		birds = append(birds, bird)
	}
	// print the length, and all the birds
	fmt.Printf("found %d birds: %+v", len(birds), birds)
}

func insertRow(db *sql.DB) {
	// sample data that we want to insert
	newBird := Bird{
		Species:     "rooster",
		Description: "wakes you up in the morning",
	}
	// the `Exec` method returns a `Result` type instead of a `Row`
	// we follow the same argument pattern to add query params
	result, err := db.Exec("INSERT INTO birds (bird, description) VALUES ($1, $2)", newBird.Species, newBird.Description)
	if err != nil {
		log.Fatalf("could not insert row: %v", err)
	}

	// the `Result` type has special methods like `RowsAffected` which returns the
	// total number of affected rows reported by the database
	// In this case, it will tell us the number of rows that were inserted using
	// the above query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("could not get affected rows: %v", err)
	}
	// we can log how many rows were inserted
	fmt.Println("inserted", rowsAffected, "rows")
}

func executePreparedStatement(db *sql.DB) {
	// 1. Prepare the statement
	stmt, err := db.Prepare("SELECT bird, description FROM birds WHERE bird = $1")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close() // Important to close prepared statements

	// 2. Execute the statement with a parameter
	var bird Bird
	err = stmt.QueryRow("eagle").Scan(&bird.Species, &bird.Description)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("result: %+v", bird)
}

func queryCancellation(db *sql.DB) {
	// create a parent context
	ctx := context.Background()
	// create a context from the parent context with a 300ms timeout
	ctx, _ = context.WithTimeout(ctx, 300*time.Millisecond)
	// The context variable is passed to the `QueryContext` method as
	// the first argument
	// the pg_sleep method is a function in Postgres that will halt for
	// the provided number of seconds. We can use this to simulate a
	// slow query
	_, err := db.QueryContext(ctx, "SELECT * from pg_sleep(1)")
	if err != nil {
		log.Fatalf("could not execute query: %v", err)
	}
}
