package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Character struct {
	EnName           string
	JpName           string
	EnAttribute      string
	JpAttribute      string
	Weapon           string
	Role             string
	Gender           string
	Race1            string
	Race2            string
	EnLeaderBuff     string
	JpLeaderBuff     string
	EnSkillName      string
	JpSkillName      string
	EnSkillDesc      string
	JpSkillDesc      string
	SkillCost        interface{}
	EnAbility1       string
	EnAbility2       string
	EnAbility3       string
	JpAbility1       string
	JpAbility2       string
	JpAbility3       string
	Rarity           int
	SkillTypeIconUrl string
	SpriteURL        string
	GifURL           string
	Nicknames        string
}

type exactCharacter struct {
	Main Character   `json:"unit"`
	Alts []Character `json:"versions"`
}

type searchMatches struct {
	Matches []Character `json:"matches"`
}

var db_url string

func main() {
	db_url = os.Getenv("DATABASE_URL")
	router := mux.NewRouter().StrictSlash(true)
	s := router.PathPrefix("/api").Subrouter()
	s.HandleFunc("/v1/lookup", lookup).Methods("GET", "OPTIONS")
	s.HandleFunc("/v2/lookup", lookup2).Methods("GET", "OPTIONS")
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
}

func lookup(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	names, ok := vals["name"]
	name := names[0]
	var char []Character

	if ok {
		db, err := gorm.Open("postgres", db_url)
		if err != nil {
			fmt.Println("failed")

			log.Fatalf("Unable to read client secret file: %v", err)
		}
		fmt.Println("connected")
		db.Where("'"+name+"'"+" = ANY(characters.nicknames)").Or("en_name LIKE ?", "%"+name+"%").Or("jp_name LIKE ?", "%"+name+"%").Find(&char)
		defer db.Close()
	}
	for i := range char {
		char[i].Nicknames = strings.ReplaceAll(char[i].Nicknames, "{", "[")
		char[i].Nicknames = strings.ReplaceAll(char[i].Nicknames, "}", "]")
	}
	enc := json.NewEncoder(w)
	enc.Encode(char)
}

func lookup2(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	names, ok := vals["name"]
	name := names[0]
	var char searchMatches
	var exactMatch exactCharacter
	exact := true
	// db_url = fmt.Sprintf("host=%s port=%d user=%s "+
	// 	"password=%s dbname=%s sslmode=%s",
	// 	host, port, user, password, dbname, sslmode)
	if ok {
		db, err := gorm.Open("postgres", db_url)

		if err != nil {
			fmt.Println("failed")

			log.Fatalf("Unable to read client secret file: %v", err)
		}
		fmt.Println("connected")

		query := `select c.*
from characters c
         LEFT JOIN character_versions cv
                   ON c.id = cv.character_id
where (cv.version_id in
       (select version_id
        from character_versions cv
                 LEFT JOIN characters c ON c.id = cv.character_id
        where en_name = '` + name + `'
           or '` + name + `' = ANY (c.nicknames)
       ));`

		db.Raw(query).Scan(&char.Matches)

		if len(char.Matches) == 0 {
			exact = false
			db.Where("'"+name+"'"+" = ANY(characters.nicknames)").Or("en_name LIKE ?", "%"+name+"%").Or("jp_name LIKE ?", "%"+name+"%").Find(&char.Matches)
		}

		defer db.Close()
	}

	for i := range char.Matches {
		char.Matches[i].Nicknames = strings.ReplaceAll(char.Matches[i].Nicknames, "{", "[")
		char.Matches[i].Nicknames = strings.ReplaceAll(char.Matches[i].Nicknames, "}", "]")

		if strings.ToLower(char.Matches[i].EnName) == name {
			exactMatch.Main = char.Matches[i]
		}
		r := regexp.MustCompile(`[^\s"',\]\[]+|'([^']*)|"([^,]*)"|'([^]]*)|'([^[]*)`)
		nicks := r.FindAllString(char.Matches[i].Nicknames, -1)
		for _, k := range nicks {
			k = strings.ReplaceAll(k, `"`, "")
			if strings.ToLower(k) == name {
				exactMatch.Main = char.Matches[i]
			}
		}
		exactMatch.Alts = append(exactMatch.Alts, char.Matches[i])
	}
	if exact {
		enc := json.NewEncoder(w)
		enc.Encode(exactMatch)
	} else {
		enc := json.NewEncoder(w)
		enc.Encode(char)
	}
}
