package main

import (
	"database/sql"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var CREATE_TABLE_SQL = `
        create table codes (path text not null primary key, password text, valueMap text);
        delete from codes;
        `

func TestGetDefaultCodeMap(t *testing.T) {
	testMap := getDefaultCodeMap()
	// TODO: test moar
	if testMap["a"] != "z" {
		t.Errorf("getDefaultCodeMap() expected z, got: %s", testMap["a"])
	}
}

func TestGetPathCodeMap(t *testing.T) {
	testDB := setupTestDB(t)

	testMap, err := getPathCodeMap(testDB, "testpath")
	if err != nil {
		t.Errorf("error in getPathCodeMap(): %s", err)
	}

	if len(testMap) != 26 {
		t.Errorf("getPathCodeMap() length expected 26, got: %v", len(testMap))
	}

	if testMap["a"] != "z" {
		t.Errorf("getPathCodeMap() expected z, got: %s", testMap["a"])
	}
}

func TestSetPathCodeMap(t *testing.T) {
	testMap := getDefaultCodeMap()
	// quick sanity check
	if testMap["a"] != "z" {
		t.Errorf("getDefaultCodeMap() expected z, got: %s", testMap["a"])
	}

	// change values for fun
	testMap["a"] = "a"
	testMap["z"] = "z"

	testDB := setupTestDB(t)

	// setupTestDB() already created testpath for us
	err := setPathCodeMap(testDB, "testpath", testMap)
	if err != nil {
		t.Errorf("error in setPathCodeMap(): %s", err)
	}

}

func TestGetPathPass(t *testing.T) {
	testDB := setupTestDB(t)
	tempPassHashed, err := getPathPass(testDB, "testpath")
	if err != nil {
		t.Errorf("error in getPathPass(): %s", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(tempPassHashed), []byte("password123"))
	if err != nil {
		t.Errorf("error in getPathPass(): %s", err)
	}

}

func TestCreateNewCode(t *testing.T) {
	testDB := setupTestDB(t)
	err := createNewCode(testDB, "testerpath2", "testpassword")
	if err != nil {
		t.Errorf("error in createNewCode(): %s", err)
	}
	stmt, err := testDB.Prepare("select path from codes where path = ?")
	if err != nil {
		t.Errorf("error in createNewCode(): %s", err)
	}
	defer stmt.Close()

	var returnedPath string
	err = stmt.QueryRow("testerpath2").Scan(&returnedPath)
	if err != nil {
		t.Errorf("error in createNewCode(): %s", err)
	}
	if returnedPath != "testerpath2" {
		t.Errorf("expected testerpath2, got: %s", returnedPath)
	}

}

func TestIsClaimed(t *testing.T) {
	// isClaimed(db *sql.DB, path string)

	testDB := setupTestDB(t)
	shouldPass := isClaimed(testDB, "testpath")
	if !shouldPass {
		t.Errorf("expected true, got: %v", shouldPass)
	}

	shouldFail := isClaimed(testDB, "doesnotcompute")
	if shouldFail {
		t.Errorf("expected false, got: %v", shouldFail)
	}
}

func setupTestDB(t *testing.T) *sql.DB {

	// setup test db
	db, err := sql.Open("sqlite3", "./testing-only.db")
	if err != nil {
		t.Errorf("unable to create testing db: %s", err)
	}

	_, err = db.Exec(CREATE_TABLE_SQL)
	if err != nil {
		t.Errorf("unable to create testing db table structure: %s", err)
	}

	// TODO: should we eat our own dogfood here?
	// insert dummy data
	createNewCode(db, "testpath", "password123")

	t.Cleanup(func() {
		db.Close()
		os.Remove("./testing-only.db")
	})

	return db
}
