package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"os"

	"github.com/gorilla/mux"
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
	Nicknames		[]string
}
var db_url string
func main() {
	db_url = os.Getenv("DATABASE_URL")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/lookup", lookup).Methods("GET", "OPTIONS")
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), router))
}

func lookup(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	names, ok := vals["name"]
	name := names[0]
	var char []Character
	fmt.Println("name:" + name)

	if ok {
		db, err := gorm.Open("postgres",db_url)
		if err != nil {
			fmt.Println("failed")

			log.Fatalf("Unable to read client secret file: %v", err)
		}
		fmt.Println("connected")
		db.Where( "'" + name + "'" +  " = ANY(characters.nicknames)").Or("en_name LIKE ?", "%" + name + "%").Or("jp_name LIKE ?", "%" + name + "%").Find(&char)
		defer db.Close()
	}
	enc := json.NewEncoder(w)
	enc.Encode(char)
}
