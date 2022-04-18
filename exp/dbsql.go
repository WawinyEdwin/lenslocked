package main

// import (
// 	"database/sql"
// 	"fmt"

// 	//using the _ tells the go compiler that we wont be using the package directly in our code but we need it to be imported.

	
// 	_ "github.com/lib/pq"
// )

// //information about our database
// const (
// 	host     = "localhost"
// 	port     = 5432
// 	user     = "postgres"
// 	password = "user"
// 	dbname   = "lenslocked_dev"
// )

// func main() {
// 	//we use this to connect to the pg server
// 	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s"+" password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

// 	//opening a database connection
// 	db, err := sql.Open("postgres", psqlInfo)
// 	if err != nil {
// 		panic(err)
// 	}

// 	var id int
// 	for i  := 1; i < 6; i++ {
// 		//create some fake data
// 		userId := 1
// 		if i > 3 {
// 			userId = 2
// 		}
// 		amount := 1000 * i
// 		description := fmt.Sprintf("USB-C Adapter x%d", i)

// 		err = db.QueryRow(`INSERT INTO orders (user_id, amount, description) VALUES ($1, $2, $3) RETURNING id`, userId, amount, description).Scan(&id)
// 		if err != nil {
// 			panic(err)
// 		}
// 		fmt.Println("Created an order with the ID:", id)
// 	}

// 	// //pinging the db
// 	// err = db.Ping()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	fmt.Println("Connection Successfull!")

// 	//inserting users into our db
// 	// _, err = db.Exec(`INSERT INTO users(name, email) VALUES($1, $2)`, "chinua achebe", "chinua@mail.com")
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	//acquiring the id of a new record.
// 	//inserting and retrieving a the ID OF A NEW RECORD
// 	// var id int
// 	// row := db.QueryRow(`INSERT INTO users(name, email) VALUES ($1, $2) RETURNING id`, "chinua achebe", "chinua@mail.com")
// 	// //Row.Scan() tell that we would like to store the data retrieved from the database
// 	// err = row.Scan(&id)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	//querying a user br ID
// 	// var id int
// 	// var name, email string
// 	// rows, err := db.Query(`SELECT id, name, email FROM users WHERE ID=$1 OR ID > $2`, "chinua@mail.com", 3)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// for rows.Next() {
// 	// 	rows.Scan(&id, &name, &email)
// 	// 	fmt.Println("ID:", id, "Name:", name, "Email:", email)
// 	// }
// 	db.Close()
// }
