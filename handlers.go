package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

type FormResponse struct {
	Path       string
	IsClaimed  bool
	ErrorMsg   string
	ValueMap   map[string]string
	EncodedVal string
	DecodedVal string
}

var templates = template.Must(template.ParseGlob("views/*.html"))

func getIndex(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templateResponse("index", FormResponse{}, w)
	})
}

func getCode(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		log.Println("getCode() id is: " + id)

		stmt, err := db.Prepare("select valueMap from codes where path = ?")
		if err != nil {
			log.Println("unable to prepare SQL statement: ", err)
			toReturnErr := FormResponse{
				Path:       id,
				ErrorMsg:   "Unable to load valueMap",
				IsClaimed:  isClaimed(db, id),
				ValueMap:   getDefaultCodeMap(),
				EncodedVal: "",
				DecodedVal: "",
			}

			templateResponse("code", toReturnErr, w)
			return

		}
		defer stmt.Close()

		var codeTableDB string
		err = stmt.QueryRow(id).Scan(&codeTableDB)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Println("no rows returned, sending default code map")
				// there were no rows, but otherwise no error occurred
				// use defalt code map
				toReturn := FormResponse{
					Path:       id,
					IsClaimed:  isClaimed(db, id),
					ValueMap:   getDefaultCodeMap(),
					EncodedVal: "",
					DecodedVal: "",
				}
				templateResponse("code", toReturn, w)
				return
			} else {
				log.Println("unable to query row: ", err)
				toReturnErr := FormResponse{
					Path:       id,
					ErrorMsg:   "Unable to load valueMap",
					IsClaimed:  isClaimed(db, id),
					ValueMap:   getDefaultCodeMap(),
					EncodedVal: "",
					DecodedVal: "",
				}
				templateResponse("code", toReturnErr, w)
				return
			}
		}

		var codeTable map[string]string
		if err = json.Unmarshal([]byte(codeTableDB), &codeTable); err != nil {
			log.Println("error unmarshaling JSON: ", err)
			toReturnErr := FormResponse{
				Path:       id,
				ErrorMsg:   "Unable to load valueMap",
				IsClaimed:  isClaimed(db, id),
				ValueMap:   getDefaultCodeMap(),
				EncodedVal: "",
				DecodedVal: "",
			}
			templateResponse("code", toReturnErr, w)
			return
		}

		toReturn := FormResponse{
			Path:       id,
			IsClaimed:  isClaimed(db, id),
			ValueMap:   codeTable,
			EncodedVal: "",
			DecodedVal: "",
		}
		templateResponse("code", toReturn, w)
	})
}

func postEncode(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		r.ParseForm()

		myMap, err := getPathCodeMap(db, id)
		if err != nil {
			toReturnErr := FormResponse{
				Path:       id,
				ErrorMsg:   "Unable to load valueMap",
				IsClaimed:  isClaimed(db, id),
				ValueMap:   getDefaultCodeMap(),
				EncodedVal: "",
				DecodedVal: "",
			}
			templateResponse("code", toReturnErr, w)
			return

		}
		toEncode := r.FormValue("encInput")
		valToReturn := ""
		if toEncode != "" {
			for _, char := range toEncode {
				if string(char) == " " {
					valToReturn += " "
				} else {
					valToReturn += myMap[string(char)]
				}
			}
		}
		toReturn := FormResponse{
			Path:       id,
			IsClaimed:  isClaimed(db, id),
			ValueMap:   myMap,
			EncodedVal: valToReturn,
			DecodedVal: "",
		}
		templateResponse("code", toReturn, w)

	})
}

func postDecode(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		r.ParseForm()

		myMap, err := getPathCodeMap(db, id)
		if err != nil {
			toReturnErr := FormResponse{
				Path:       id,
				ErrorMsg:   "Unable to load valueMap",
				IsClaimed:  isClaimed(db, id),
				ValueMap:   getDefaultCodeMap(),
				EncodedVal: "",
				DecodedVal: "",
			}
			templateResponse("code", toReturnErr, w)
			return

		}
		toDecode := r.FormValue("decInput")
		valToReturn := ""

		if toDecode != "" {
			for _, char := range toDecode {
				if string(char) == " " {
					valToReturn += " "
				} else {
					for k, v := range myMap {
						if v == string(char) {
							valToReturn += k
						}
					}
				}
			}

		}
		toReturn := FormResponse{
			Path:       id,
			IsClaimed:  isClaimed(db, id),
			ValueMap:   myMap,
			EncodedVal: "",
			DecodedVal: valToReturn,
		}
		templateResponse("code", toReturn, w)

	})
}

func postSaveMap(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		r.ParseForm()

		pathPass := r.FormValue("pathPass")
		currentPathPass, err := getPathPass(db, id)
		if err != nil {
			log.Println(err)
			myMap, err := getPathCodeMap(db, id)
			if err != nil {
				log.Println(err)
				myMap = getDefaultCodeMap()
			}
			toReturnErr := FormResponse{
				Path:       id,
				ErrorMsg:   "Unable to verify secret",
				IsClaimed:  isClaimed(db, id),
				ValueMap:   myMap,
				EncodedVal: "",
				DecodedVal: "",
			}
			templateResponse("code", toReturnErr, w)
			return

		}

		if currentPathPass == "" {
			// path has not been claimed, set initial pass
			createNewCode(db, id, strings.TrimSpace(pathPass))
		} else {

			if !comparePasswords(currentPathPass, pathPass) {
				myMap, err := getPathCodeMap(db, id)
				if err != nil {
					myMap = getDefaultCodeMap()
				}

				toReturnErr := FormResponse{
					Path:       id,
					IsClaimed:  isClaimed(db, id),
					ErrorMsg:   "Invalid Secret",
					ValueMap:   myMap,
					EncodedVal: "",
					DecodedVal: "",
				}
				templateResponse("code", toReturnErr, w)
				return

			}
		}

		myMap, err := getPathCodeMap(db, id)
		if err != nil {
			myMap = getDefaultCodeMap()

			toReturnErr := FormResponse{
				Path:       id,
				IsClaimed:  isClaimed(db, id),
				ErrorMsg:   "Invalid Secret",
				ValueMap:   myMap,
				EncodedVal: "",
				DecodedVal: "",
			}
			templateResponse("code", toReturnErr, w)
			return
		}

		for k, _ := range myMap {
			myMap[k] = r.FormValue(k)
		}

		setPathCodeMap(db, id, myMap)

		toReturn := FormResponse{
			Path:       id,
			IsClaimed:  isClaimed(db, id),
			ValueMap:   myMap,
			EncodedVal: "",
			DecodedVal: "",
		}
		templateResponse("code", toReturn, w)

	})
}

func templateResponse(templateName string, pageBody FormResponse, w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, templateName+".html", pageBody)

	if err != nil {
		log.Println("unable to execute template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
