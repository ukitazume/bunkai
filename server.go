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
	Id        int64
	Email     string
	Password  string
	CreatedAt int64
	UpdatedAt int64
}

func newUser(email, password string) User {
	return User{
		Email:     email,
		Password:  password,
		CreatedAt: time.Now().UnixNano(),
	}
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
	m.Post("/sentences", SentenceCreate)
	m.Get("/sentences", SentenceList)
	m.Delete("/sentences/:id", SentenceDelete)

	m.Get("/users/:id", UserGet)
	m.Post("/users", UserCreate)

	m.Run()
}

func Home(ren render.Render) {
	ren.HTML(200, "home", nil)
}

func SentenceCreate(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
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

func SentenceList(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
	var sens []Sentence
	_, err := dbmap.Select(&sens, "select * from sentences order by id")
	PanicIf(err)
	ren.JSON(200, sens)
}

func SentenceDelete(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	_, err := dbmap.Exec("DELETE FROM sentences WHERE id= $1", params["id"])
	PanicIf(err)
	ren.JSON(200, nil)
}

func UserGet(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	var usr User
	log.Println(params)
	err := dbmap.SelectOne(&usr, "SELECT * from users WHERE id = $1", params["id"])
	PanicIf(err)
	ren.JSON(200, usr)
}

func UserCreate(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
	usr := newUser(req.FormValue("email"), req.FormValue("password"))
	err := dbmap.Insert(&usr)
	PanicIf(err)
	ren.JSON(200, usr)
}
