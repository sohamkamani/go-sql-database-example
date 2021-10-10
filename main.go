// file: main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	// we have to import the driver, but don't use it in our code
	// so we use the `_` symbol
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Bird struct {
	Species     string
	Description string
}

func main() {
	// The `sql.Open` function opens a new `*sql.DB` instance. We specify the driver name
	// and the URI for our database. Here, we're using a Postgres URI
	db, err := sql.Open("pgx", "postgresql://localhost:5432/bird_encyclopedia")
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
	// method. If no error is returned, we can assume a successful connection
	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Println("database is reachable")

	// ctx := context.Background()
	// conn, err := db.Conn(ctx)
	// if err != nil {
	// 	log.Fatalf("could not get connection: %v", err)
	// }

	// `QueryRow` always returns a single row from the database
	birdName := "eagle"
	// For Postgres, parameters are specified using the "$" symbol, along with the index of
	// the param. Variables should be added as arguments in the same order
	// The sql library takes care of converting types from Go to SQL based on the driver
	row := db.QueryRow("SELECT bird, description FROM birds WHERE bird = $1 LIMIT $2", birdName, 1)

	// the code to scan the obtained row is the same as before
	//...

	// Create a new `Bird` instance to hold our query results
	bird := Bird{}
	// the retrieved columns in our row are written to the provided addresses
	// the arguments should be in the same order as the columns defined in
	// our query
	if err := row.Scan(&bird.Species, &bird.Description); err != nil {
		log.Fatalf("could not scan row: %v", err)
	}
	fmt.Printf("found bird: %+v\n", bird)

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
	fmt.Printf("found %d birds: %+v\n", len(birds), birds)

	// this is just to remove the inserted row from the previous run
	_, err = db.Exec("DELETE FROM birds WHERE bird=$1", "rooster")
	if err != nil {
		log.Fatalf("could not delete row: %v", err)
	}

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

	// create a parent context
	ctx := context.Background()
	// create a context from the parent context with a 300ms timeout
	ctx, _ = context.WithTimeout(ctx, 300*time.Millisecond)
	// The context variable is passed to the `QueryContext` method as
	// the first argument
	_, err = db.QueryContext(ctx, "SELECT * from pg_sleep(1)")
	if err != nil {
		log.Fatalf("could not execute query: %v", err)
	}

}
