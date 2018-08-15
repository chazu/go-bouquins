package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"

	"meutel.net/meutel/go-bouquins/bouquins"
)

// ReadConfig loads configuration file and initialize default value
func ReadConfig() (*bouquins.Conf, error) {
	conf := new(bouquins.Conf)
	confPath := "bouquins.json"
	if len(os.Args) > 1 {
		confPath = os.Args[1]
	}
	confFile, err := os.Open(confPath)
	if err == nil {
		defer confFile.Close()
		err = json.NewDecoder(confFile).Decode(conf)
	} else {
		log.Println("no conf file, using defaults")
		err = nil
	}
	// default values
	if conf.CalibrePath == "" {
		conf.CalibrePath = "."
	}
	if conf.DbPath == "" {
		conf.DbPath = conf.CalibrePath + "/metadata.db"
	}
	if conf.UserDbPath == "" {
		conf.UserDbPath = "./users.db"
	}
	if conf.BindAddress == "" {
		conf.BindAddress = ":9000"
	}
	return conf, err
}

func initApp() *bouquins.Bouquins {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	conf, err := ReadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	tpl, err := bouquins.TemplatesFunc(conf.Prod).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalln(err)
	}
	db, err := sql.Open("sqlite3", conf.DbPath)
	if err != nil {
		log.Fatalln(err)
	}
	userdb, err := sql.Open("sqlite3", conf.UserDbPath)
	if err != nil {
		log.Fatalln(err)
	}

	app := &bouquins.Bouquins{
		Tpl:       tpl,
		DB:        db,
		UserDB:    userdb,
		Conf:      conf,
		OAuthConf: make(map[string]*oauth2.Config),
		Cookies:   sessions.NewCookieStore([]byte(conf.CookieSecret)),
	}
	for _, provider := range bouquins.Providers {
		app.OAuthConf[provider.Name()] = provider.Config(conf)
	}
	err = app.PrepareAll()
	if err != nil {
		log.Fatalln(err)
	}
	router(app)
	return app
}

func assets(calibre string) {
	http.Handle(bouquins.URLJs, http.StripPrefix("/"+bouquins.Version, http.FileServer(http.Dir("assets"))))
	http.Handle(bouquins.URLCss, http.StripPrefix("/"+bouquins.Version, http.FileServer(http.Dir("assets"))))
	http.Handle(bouquins.URLFonts, http.StripPrefix("/"+bouquins.Version, http.FileServer(http.Dir("assets"))))
}

func handle(f func(res http.ResponseWriter, req *http.Request) error) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		err := f(res, req)
		if err != nil {
			log.Println(err)
			http.Error(res, err.Error(), 500)
		}
	}
}

func handleURL(url string, f func(res http.ResponseWriter, req *http.Request) error) {
	http.HandleFunc(url, handle(f))
}

func router(app *bouquins.Bouquins) {
	assets(app.Conf.CalibrePath)
	http.Handle(bouquins.URLCalibre, app.CalibreFileServer())
	handleURL(bouquins.URLIndex, app.IndexPage)
	handleURL(bouquins.URLLogin, app.LoginPage)
	handleURL(bouquins.URLLogout, app.LogoutPage)
	handleURL(bouquins.URLCallback, app.CallbackPage)
	handleURL(bouquins.URLBooks, app.BooksPage)
	handleURL(bouquins.URLAuthors, app.AuthorsPage)
	handleURL(bouquins.URLSeries, app.SeriesPage)
	handleURL(bouquins.URLSearch, app.SearchPage)
	handleURL(bouquins.URLAbout, app.AboutPage)
}

func main() {
	app := initApp()
	defer app.DB.Close()
	defer app.UserDB.Close()
	http.ListenAndServe(app.Conf.BindAddress, nil)
}
