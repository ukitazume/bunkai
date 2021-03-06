package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
)

const (
	CookieSecret       = "secretedesse"
	DBName             = "bunkai"
	BowerComponentPath = "./client/src/components/"
)

func SetupDB() *gorp.DbMap {
	db, err := sql.Open("postgres", "dbname="+DBName+" sslmode=disable")
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

func Md5(org string) string {
	hasher := md5.New()
	hasher.Write([]byte(org))
	return hex.EncodeToString(hasher.Sum(nil))
}

type Sentence struct {
	Id        int64
	UserId    int64
	Text      string
	Url       string
	CreatedAt int64
}

func newSentence(usr User, text, url string) Sentence {
	return Sentence{
		UserId:    usr.Id,
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
		Password:  Md5(password),
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

	store := sessions.NewCookieStore([]byte(CookieSecret))
	m.Use(sessions.Sessions("bunkaisession", store))
	log.Println("env is", martini.Env)
	m.Map(martini.Env)

	m.Group("/", func(m martini.Router) {
		m.Get("", Home)
	}, AssetMap)

	m.Group("/api", func(m martini.Router) {
		m.Post("/login", PostLogin)
		m.Post("/users", UserCreate)

		m.Group("/sentences", func(m martini.Router) {
			m.Post("", SentenceCreate)
			m.Get("", SentenceList)
			m.Delete("/:id", SentenceDelete)
		}, RequireLogin)

		m.Group("/users", func(m martini.Router) {
			m.Post("/logout", Logout)
			m.Get("/me", UserGet)
		}, RequireLogin)
	})

	m.Run()
}

func AssetMap(c martini.Context) {
	c.Map(JsComponent())
}

func JsComponent() []string {
	jsdirs, err := ioutil.ReadDir(BowerComponentPath)
	PanicIf(err)
	var jslibNamePath []string
	for _, value := range jsdirs {
		s := []string{BowerComponentPath, value.Name(), "/", value.Name(), ".js"}
		jslibNamePath = append(jslibNamePath, strings.Join(s, ""))
	}
	return jslibNamePath
}

func RequireLogin(ren render.Render, req *http.Request, s sessions.Session, dbmap *gorp.DbMap, c martini.Context) {

	var usr User
	err := dbmap.SelectOne(&usr, "SELECT * from users WHERE id = $1", s.Get("userId"))

	if err != nil {
		ren.JSON(http.StatusForbidden, nil)
		return
	}

	c.Map(usr)
}

func Logout(ren render.Render, req *http.Request, s sessions.Session) {
	s.Delete("userId")
	ren.JSON(http.StatusAccepted, nil)
}

func PostLogin(req *http.Request, dbmap *gorp.DbMap, s sessions.Session) (int, string) {
	var userId string

	email, password := req.FormValue("email"), req.FormValue("password")
	err := dbmap.SelectOne(&userId, "select id from users where email=$1 and password=$2", email, Md5(password))

	if err != nil {
		return 401, "Unauthorized"
	}

	s.Set("userId", userId)

	return 200, "User id is " + userId
}

func Home(ren render.Render) {
	ren.HTML(200, "home", nil)
}

func SentenceCreate(ren render.Render, req *http.Request, dbmap *gorp.DbMap, usr User) {
	sentence := newSentence(usr, req.FormValue("text"), req.FormValue("url"))
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

func SentenceList(ren render.Render, req *http.Request, dbmap *gorp.DbMap, usr User) {
	var sens []Sentence
	_, err := dbmap.Select(&sens, "select * from sentences WHERE userid = $1 order by id", usr.Id)
	PanicIf(err)
	ren.JSON(200, sens)
}

func SentenceDelete(ren render.Render, params martini.Params, dbmap *gorp.DbMap) {
	_, err := dbmap.Exec("DELETE FROM sentences WHERE id= $1", params["id"])
	PanicIf(err)
	ren.JSON(200, nil)
}

func UserGet(ren render.Render, params martini.Params, dbmap *gorp.DbMap, s sessions.Session) {
	var usr User
	log.Println(params)
	err := dbmap.SelectOne(&usr, "SELECT * from users WHERE id = $1", s.Get("userId"))
	PanicIf(err)
	ren.JSON(200, usr)
}

func UserCreate(ren render.Render, req *http.Request, dbmap *gorp.DbMap) {
	usr := newUser(req.FormValue("email"), req.FormValue("password"))
	err := dbmap.Insert(&usr)
	PanicIf(err)
	ren.JSON(200, usr)
}
