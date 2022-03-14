// Copyright 2020 Authors of Arktos.
// Copyright 2020 Authors of Arktos - file modified.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"database/sql"
	"fmt"
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/iam/model"
	"log"
	"os"
)

// CreateConnection creates connection with postgres db
func CreateConnection() *sql.DB {

	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")
	POSTGRES_USER := os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD := os.Getenv("POSTGRES_PASSWORD")
	POSTGRES_DB := os.Getenv("POSTGRES_DB")

	// Create connection string
	connStr := "host=" + DB_HOST + " port=" + DB_PORT + " dbname=" + POSTGRES_DB + " user=" + POSTGRES_USER + " password=" + POSTGRES_PASSWORD + " sslmode=disable"

	// Open the connection
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	// check the connection
	err = db.Ping()

	if err != nil {
		log.Fatalf("Error connecting to the database: %s", err)
	}

	fmt.Println("Successfully connected!")
	// return the connection
	return db
}

// insert one user in the DB

func InsertUser(user model.User) int64 {

	// create the postgres db connection
	db := CreateConnection()

	// close the db connection
	defer db.Close()

	// create the insert sql query
	// returning userid will return the id of the inserted user
	sqlStatement := `INSERT INTO userdetails (username, password, token, type, tenant, role, creationtime, namespace) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT ON CONSTRAINT userdetails_username_key DO UPDATE SET token=EXCLUDED.token RETURNING userid;`
	// the inserted id will store in this id
	var id int64

	// execute the sql statement
	// Scan function will save the insert id in the id
	err := db.QueryRow(sqlStatement, user.Username, user.Password, user.Token, user.Type, user.Tenant, user.Role, user.CreationTimestamp, user.NameSpace).Scan(&id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	fmt.Printf("Inserted a single record %v", id)
	// return the inserted id
	return id
}

// GetUser gets a user from the DB by its userid
func GetUser(param string) (model.UserDetails, error) {
	// create the postgres db connection
	db := CreateConnection()

	// close the db connection
	defer db.Close()

	// create a user of model.User type
	var user model.UserDetails

	user.Phase = "Active"
	user.TypeMeta.Kind = "User"

	// create the select sql query
	sqlStatement := `SELECT * FROM userdetails WHERE username=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, param)

	// unmarshal the row object to user
	err := row.Scan(&user.ObjectMeta.ID, &user.ObjectMeta.Username, &user.ObjectMeta.Password, &user.ObjectMeta.Token, &user.ObjectMeta.Type, &user.ObjectMeta.Tenant, &user.ObjectMeta.Role, &user.ObjectMeta.CreationTimestamp, &user.ObjectMeta.NameSpace)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return user, err
}

func GetUserDetail(param string) (model.UserDetails, error) {
	// create the postgres db connection
	db := CreateConnection()

	// close the db connection
	defer db.Close()

	// create a user of model.User type
	var user model.UserDetails

	user.Phase = "Active"
	user.TypeMeta.Kind = "User"

	// create the select sql query
	sqlStatement := `SELECT * FROM userdetails WHERE username=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, param)

	// unmarshal the row object to user
	err := row.Scan(&user.ObjectMeta.ID, &user.ObjectMeta.Username, &user.ObjectMeta.Password, &user.ObjectMeta.Token, &user.ObjectMeta.Type, &user.ObjectMeta.Tenant, &user.ObjectMeta.Role, &user.ObjectMeta.CreationTimestamp, &user.ObjectMeta.NameSpace)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return user, err
}

// GetAllUsers gets all the users from the DB
func GetAllUsers(tenant string) (*model.UserList, error) {
	// create the postgres db connection
	db := CreateConnection()

	// close the db connection
	defer db.Close()

	userList := new(model.UserList)

	// create the select sql query
	rows := new(sql.Rows)
	var err error
	if tenant == "system" || tenant == "" {
		sqlStatement := `SELECT * FROM userdetails`

		// execute the sql statement
		rows, err = db.Query(sqlStatement)

		if err != nil {
			log.Fatalf("Unable to execute the query. %v", err)
		}
	} else {
		sqlStatement := `SELECT * FROM userdetails WHERE tenant=$1`

		// execute the sql statement
		rows, err = db.Query(sqlStatement, tenant)

		if err != nil {
			log.Fatalf("Unable to execute the query. %v", err)
		}
	}

	// close the statement
	defer rows.Close()

	// iterate over the rows
	count := 0
	for rows.Next() {
		var user model.UserDetails

		// unmarshal the row object to user
		err = rows.Scan(&user.ObjectMeta.ID, &user.ObjectMeta.Username, &user.ObjectMeta.Password, &user.ObjectMeta.Token, &user.ObjectMeta.Type, &user.ObjectMeta.Tenant, &user.ObjectMeta.Role, &user.ObjectMeta.CreationTimestamp, &user.ObjectMeta.NameSpace)

		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}
		user.Phase = "Active"
		user.TypeMeta.Kind = "User"

		// append the user in the users slice
		userList.Users = append(userList.Users, user)
		count++

	}
	userList.ListMeta = api.ListMeta{TotalItems: count}
	// return empty user on error
	return userList, err
}

// DeleteUser deletes a user from database
func DeleteUser(id int64) int64 {

	// create the postgres db connection
	db := CreateConnection()

	// close the db connection
	defer db.Close()

	// create the delete sql query
	sqlStatement := `DELETE FROM userdetails WHERE userid=$1`

	// execute the sql statement
	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// check how many rows affected
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

func DeleteTenantUser(tenant string) int64 {

	// create the postgres db connection
	db := CreateConnection()

	// close the db connection
	defer db.Close()

	// create the delete sql query
	sqlStatement := `DELETE FROM userdetails WHERE tenant=$1`

	// execute the sql statement
	res, err := db.Exec(sqlStatement, tenant)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// check how many rows affected
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}
