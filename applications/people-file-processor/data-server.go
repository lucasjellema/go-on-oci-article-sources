package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var autonomousDB = map[string]string{
	"service":        "k8j2fvxbaujdcfy_goonocidb_medium.adb.oraclecloud.com",
	"username":       "demo",
	"server":         "adb.us-ashburn-1.oraclecloud.com",
	"port":           "1522",
	"password":       "thePassword1",
	"walletLocation": ".",
}

type Person struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	JuicyDetails string `json:"comment"`
}

const (
	PEOPLE_TABLE_NAME = "PEOPLE"
)

var database *sql.DB

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var tblCount int32
	err := db.QueryRow("SELECT count(table_name) tbl_count FROM user_tables where table_name = upper(:tablename)", tableName).Scan(&tblCount)
	exists := tblCount > 0
	return exists, err
}

func InitializeDataServer(db *sql.DB) error {
	database = db
	exists, err := tableExists(db, PEOPLE_TABLE_NAME)
	if err != nil {
		return err
	}
	if !exists {
		_, err = db.Exec(createTableStatement)
		log.Printf("Created table %s in Oracle Database \n", PEOPLE_TABLE_NAME)
		if err != nil {
			return err
		}
	}
	return nil
}

func PeopleJSONProcessor(peopleJson []byte) {
	var persons []Person
	json.Unmarshal(peopleJson, &persons)
	for _, person := range persons {
		persistPerson(person)
	}
}

func persistPerson(person Person) {
	ctx := context.Background()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = mergePerson(ctx, tx, person)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Merged record in table %s for person %s", PEOPLE_TABLE_NAME, person.Name)
	}
}

const createTableStatement = "CREATE TABLE PEOPLE ( NAME VARCHAR2(100), AGE NUMBER(3), DESCRIPTION VARCHAR2(1000), CREATION_TIME TIMESTAMP DEFAULT SYSTIMESTAMP)"

func mergePerson(ctx context.Context, tx *sql.Tx, person Person) error {
	mergeStatement := fmt.Sprintf(
		`MERGE INTO %s t using (select :name name, :age age, :description description from dual) person
		ON (t.name = person.name )
		WHEN MATCHED THEN UPDATE SET age = person.age, description = person.description
		WHEN NOT MATCHED THEN INSERT (t.name, t.age, t.description) values (person.name, person.age, person.description) `,
		PEOPLE_TABLE_NAME)
	_, err := tx.ExecContext(ctx, mergeStatement, person.Name, person.Age, person.JuicyDetails)
	return err
}

func DataHandler(response http.ResponseWriter, request *http.Request) {
	log.Printf("Handle Request for method %s on path %s", request.Method, request.URL.Path)
	if request.Method == "GET" {
		queryNameParameter := request.URL.Query().Get("name")
		selectStatement := fmt.Sprintf(
			`select age, creation_time, description from %s where name = :name `,
			PEOPLE_TABLE_NAME)
		var creationTime time.Time
		var person Person
		person.Name = queryNameParameter
		row := database.QueryRow(selectStatement, person.Name)
		err := row.Scan(&person.Age, &creationTime, &person.JuicyDetails)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		personJson, _ := json.Marshal(person)
		fmt.Fprint(response, string(personJson))
	}
	if request.Method == "PUT" || request.Method == "POST" {
		var person Person
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(request.Body).Decode(&person)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		persistPerson(person)
		fmt.Fprint(response, fmt.Sprintf("Persisted %s!", person.Name))
	}
	if request.Method == "DELETE" {
		var person Person
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(request.Body).Decode(&person)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		unpersistPerson(person.Name)
		fmt.Fprint(response, fmt.Sprintf("Removed record for %s!", person.Name))
	}

}
func unpersistPerson(name string) {
	ctx := context.Background()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = deletePerson(ctx, tx, name)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Delete record from table %s for person %s", PEOPLE_TABLE_NAME, name)
	}
}
func deletePerson(ctx context.Context, tx *sql.Tx, name string) error {
	deleteStatement := fmt.Sprintf(
		`delete %s where name = :name `,
		PEOPLE_TABLE_NAME)
	_, err := tx.ExecContext(ctx, deleteStatement, name)
	return err
}
