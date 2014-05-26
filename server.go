package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
)

func SetupDB() *gorp.DbMap {
	db, err := sql.Open("postgres", "dbname=bunkai sslmode=disable")
	PanicIf(err)

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(Sentence{}, "sentences").SetKeys(true, "Id")
	dbmap.AddTableWithName(User{}, "users").SetKeys(true, "Id")

	err = dbmap.CreateTablesIfNotExists()
	PanicIf(err)

	return dbmap
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

type Sentence struct {
	Id        int64
	UserId    int64
	Text      string
	Url       string
	CreatedAt int64
}

func newSentence(text, url string) Sentence {
	return Sentence{
		Text:      text,
		Url:       url,
		CreatedAt: time.Now().UnixNano(),
	}
}

type User struct {
	Id        int64 `db:"post_id"`
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
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

func main() {
	m := martini.Classic()
	m.Map(SetupDB())
	m.Use(martini.Logger())
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))
	log.Println("env is", martini.Env)

	m.Get("/", Home)
	m.Post("/sentences", Create)
	m.Get("/sentences", List)

	m.Run()
}

func Home(ren render.Render) {
	ren.HTML(200, "home", nil)
}

func Create(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
	sentence := newSentence(req.FormValue("text"), req.FormValue("url"))
	_, err := sentence.Validate()

	if err != nil {
		msg := make(map[string]string)
		msg["error"] = err.Error()
		ren.JSON(400, msg)
	} else {
		err = dbmap.Insert(&sentence)
		PanicIf(err)
		ren.JSON(200, sentence)
	}
}

func List(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
	var sens []Sentence
	_, err := dbmap.Select(&sens, "select * from sentences order by id")
	PanicIf(err)
	ren.JSON(200, sens)
}
