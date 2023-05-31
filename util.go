package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// unicode FTW
func getDefaultCodeMap() map[string]string {
	myMap := make(map[string]string)
	for r := 'a'; r <= 'z'; r++ {
		keyString := fmt.Sprintf("%c", r)
		valueString := fmt.Sprintf("%c", 219-r)
		myMap[keyString] = valueString
	}

	return myMap
}

func getPathCodeMap(db *sql.DB, path string) (map[string]string, error) {
	stmt, err := db.Prepare("select valueMap from codes where path = ?")
	if err != nil {
		return make(map[string]string), err
	}
	defer stmt.Close()

	var codeTableDB string
	err = stmt.QueryRow(path).Scan(&codeTableDB)
	if err != nil {
		return make(map[string]string), err
	}

	var codeTable map[string]string
	if err = json.Unmarshal([]byte(codeTableDB), &codeTable); err != nil {
		return make(map[string]string), err
	}

	return codeTable, nil
}

func setPathCodeMap(db *sql.DB, path string, newMap map[string]string) error {
	b, err := json.Marshal(newMap)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("update codes set valueMap = ? where path = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(string(b), path); err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func getPathPass(db *sql.DB, path string) (string, error) {
	stmt, err := db.Prepare("select password from codes where path = ?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var currentPathPass string
	err = stmt.QueryRow(path).Scan(&currentPathPass)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("no rows returned, sending default code map")
			return "", nil
		}
		return "", err
	}

	return currentPathPass, nil
}

func createNewCode(db *sql.DB, path string, pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("insert into codes(path, password, valueMap) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	b, err := json.Marshal(getDefaultCodeMap())
	if err != nil {
		return err
	}
	_, err = stmt.Exec(path, string(hash), string(b))
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func isClaimed(db *sql.DB, path string) bool {
	currentPass, _ := getPathPass(db, path)
	if currentPass != "" {
		return true
	}
	return false
}

func comparePasswords(hashedPwd string, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
