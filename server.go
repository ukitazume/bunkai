package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
)

func SetupDB() *sql.DB {
	db, err := sql.Open("postgres", "dbname=bunkai sslmode=disable")
	PanicIf(err)

	return db
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

type Sentence struct {
	Text      string
	Url       string
	CreatedAt time.Time
}

func (sen *Sentence) Validate() (bool, error) {
	matchText, _ := regexp.MatchString(`.+`, sen.Text)
	if !matchText {
		err := errors.New("Text is invalid")
		return false, err
	}

	matchUrl, _ := regexp.MatchString(`http(s)?://([\w-]+\.)+[\w-]+(/[\w- ./?%&=]*)?`, sen.Url)
	if !matchUrl {
		err := errors.New("Url is invalid")
		return false, err
	}
	return true, nil
}

func (sen *Sentence) Save(db *sql.DB) error {
	_, verr := sen.Validate()

	rows, err := db.Query("INSERT INTO sentences (text, url, created_at) VALUES ($1, $2, $3)",
		sen.Text,
		sen.Url,
		sen.CreatedAt)
	PanicIf(err)
	defer rows.Close()

	return verr
}

func SentenceList(db *sql.DB, limit int) []Sentence {
	rows, err := db.Query("select text, url, created_at from sentences LIMIT $1", limit)
	PanicIf(err)
	s := make([]Sentence, limit)
	for rows.Next() {
		var text, url string
		var created_at time.Time
		err := rows.Scan(&text, &url, &created_at)
		PanicIf(err)
		s = append(s, Sentence{text, url, created_at})
	}
	return s
}

func main() {
	m := martini.Classic()
	m.Map(SetupDB())
	m.Use(martini.Logger())
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))
	log.Println("env is", martini.Env)

	m.Get("/", Home)
	m.Post("/create", Create)
	m.Get("/list", List)

	m.Run()
}

func Home(ren render.Render) {
	ren.HTML(200, "home", nil)
}

func Create(ren render.Render, req *http.Request, db *sql.DB) {
	sentence := Sentence{req.FormValue("text"), req.FormValue("url"), time.Now()}
	err := sentence.Save(db)
	if err != nil {
		msg := make(map[string]string)
		msg["error"] = err.Error()
		ren.JSON(400, msg)
	} else {
		ren.JSON(200, sentence)
	}

}

func List(ren render.Render, req *http.Request, db *sql.DB) {
	s := SentenceList(db, 100)
	ren.JSON(200, s)
}
