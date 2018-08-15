package bouquins

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"github.com/c2h5oh/datasize"
	"github.com/gorilla/sessions"
)

const (
	// Version defines application version
	Version = "0.1.0"

	tplBooks    = "book.html"
	tplAuthors  = "author.html"
	tplSeries   = "series.html"
	tplIndex    = "index.html"
	tplSearch   = "search.html"
	tplAbout    = "about.html"
	tplProvider = "provider.html"

	pList    = "list"
	pOrder   = "order"
	pSort    = "sort"
	pPage    = "page"
	pPerPage = "perpage"
	pTerm    = "term"

	// URLIndex url of index page
	URLIndex = "/"
	// URLLogin url of login page (OAuth 2)
	URLLogin = "/login"
	// URLLogout url of logout page
	URLLogout = "/logout"
	// URLCallback url of OAuth callback
	URLCallback = "/callback"
	// URLBooks url of books page
	URLBooks = "/books/"
	// URLAuthors url of authors page
	URLAuthors = "/authors/"
	// URLSeries url of series page
	URLSeries = "/series/"
	// URLSearch url of search page
	URLSearch = "/search/"
	// URLAbout url of about page
	URLAbout = "/about/"
	// URLJs url of js assets
	URLJs = "/" + Version + "/js/"
	// URLCss url of css assets
	URLCss = "/" + Version + "/css/"
	// URLFonts url of fonts assets
	URLFonts = "/" + Version + "/fonts/"
	// URLCalibre url of calibre resources (covers, ebooks files)
	URLCalibre = "/calibre/"
)

// UnprotectedCalibreSuffix lists suffixe of calibre file not protected by auth
var UnprotectedCalibreSuffix = [1]string{"jpg"}

// Conf App configuration
type Conf struct {
	BindAddress   string         `json:"bind-address"`
	DbPath        string         `json:"db-path"`
	CalibrePath   string         `json:"calibre-path"`
	Prod          bool           `json:"prod"`
	UserDbPath    string         `json:"user-db-path"`
	CookieSecret  string         `json:"cookie-secret"`
	ExternalURL   string         `json:"external-url"`
	ProvidersConf []ProviderConf `json:"providers"`
}

// ProviderConf OAuth2 provider configuration
type ProviderConf struct {
	Name         string `json:"name"`
	ClientID     string `json:"client-id"`
	ClientSecret string `json:"client-secret"`
}

// Bouquins contains application common resources: templates, database
type Bouquins struct {
	Tpl *template.Template
	*sql.DB
	UserDB *sql.DB
	*Conf
	OAuthConf map[string]*oauth2.Config
	Cookies   *sessions.CookieStore
}

// UserAccount is an user account
type UserAccount struct {
	ID          string // UUID
	DisplayName string
}

// Series is a book series.
type Series struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Book contains basic data on book
type Book struct {
	ID          int64   `json:"id,omitempty"`
	Title       string  `json:"title,omitempty"`
	SeriesIndex float64 `json:"series_idx,omitempty"`
	Series      *Series `json:"series,omitempty"`
}

// Author contains basic data on author
type Author struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// AuthorAdv extends Author with number of books
type AuthorAdv struct {
	Author
	Count int `json:"count,omitempty"`
}

// BookData contains data for dowloadable book
type BookData struct {
	Size   int64  `json:"size,omitempty"`
	Format string `json:"format,omitempty"`
	Name   string `json:"name,omitempty"`
}

// BookAdv extends Book with authors and tags
type BookAdv struct {
	Book
	Authors []*Author `json:"authors,omitempty"`
	Tags    []string  `json:"tags,omitempty"`
}

// AuthorFull extends Author with books, series and co-authors
type AuthorFull struct {
	Author
	Books     []*Book   `json:"books,omitempty"`
	Series    []*Series `json:"series,omitempty"`
	CoAuthors []*Author `json:"coauthors,omitempty"`
}

// BookFull extends BookAdv with all available data
type BookFull struct {
	BookAdv
	Data      []*BookData `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
	Pubdate   int64       `json:"pubdate,omitempty"`
	Isbn      string      `json:"isbn,omitempty"`
	Lccn      string      `json:"lccn,omitempty"`
	Path      string      `json:"path,omitempty"`
	UUID      string      `json:"uuid,omitempty"`
	HasCover  bool        `json:"has_cover,omitempty"`
	Lang      string      `json:"lang,omitempty"`
	Publisher string      `json:"publisher,omitempty"`
}

// SeriesAdv extends Series with count of books and authors
type SeriesAdv struct {
	Series
	Count   int64     `json:"count,omitempty"`
	Authors []*Author `json:"authors,omitempty"`
}

// SeriesFull extends SeriesAdv with related books
type SeriesFull struct {
	SeriesAdv
	Books []*Book `json:"books,omitempty"`
}

// Model is basic page model
type Model struct {
	Title    string
	Page     string
	Version  string
	Username string
}

// NewModel constructor for Model
func (app *Bouquins) NewModel(title, page string, req *http.Request) *Model {
	return &Model{
		Title:    title,
		Page:     page,
		Version:  Version,
		Username: app.Username(req),
	}
}

// IndexModel is the model for index page
type IndexModel struct {
	Model
	BooksCount int64 `json:"count"`
}

// NewIndexModel constructor IndexModel
func (app *Bouquins) NewIndexModel(title string, count int64, req *http.Request) *IndexModel {
	return &IndexModel{*app.NewModel(title, "index", req), count}
}

// NewSearchModel constuctor for search page
func (app *Bouquins) NewSearchModel(req *http.Request) *Model {
	return app.NewModel("Recherche", "search", req)
}

// ResultsModel is a generic model for list pages
type ResultsModel struct {
	Type         string `json:"type,omitempty"`
	More         bool   `json:"more"`
	CountResults int    `json:"count,omitempty"`
}

// BooksResultsModel is the model for list of books
type BooksResultsModel struct {
	ResultsModel
	Results []*BookAdv `json:"results,omitempty"`
}

// NewBooksResultsModel constuctor for BooksResultsModel
func NewBooksResultsModel(books []*BookAdv, more bool, count int) *BooksResultsModel {
	return &BooksResultsModel{ResultsModel{"books", more, count}, books}
}

// AuthorsResultsModel is the model for list of authors
type AuthorsResultsModel struct {
	ResultsModel
	Results []*AuthorAdv `json:"results,omitempty"`
}

// NewAuthorsResultsModel constuctor for AuthorsResultsModel
func NewAuthorsResultsModel(authors []*AuthorAdv, more bool, count int) *AuthorsResultsModel {
	return &AuthorsResultsModel{ResultsModel{"authors", more, count}, authors}
}

// SeriesResultsModel is the model for list of series
type SeriesResultsModel struct {
	ResultsModel
	Results []*SeriesAdv `json:"results,omitempty"`
}

// NewSeriesResultsModel constuctor for SeriesResultsModel
func NewSeriesResultsModel(series []*SeriesAdv, more bool, count int) *SeriesResultsModel {
	return &SeriesResultsModel{ResultsModel{"series", more, count}, series}
}

// BookModel is the model for single book page
type BookModel struct {
	Model
	*BookFull
}

// SeriesModel is the model for single series page
type SeriesModel struct {
	Model
	*SeriesFull
}

// AuthorModel is the model for single author page
type AuthorModel struct {
	Model
	*AuthorFull
}

// ReqParams contains request parameters for searches and lists
type ReqParams struct {
	Limit    int
	Offset   int
	Sort     string
	Order    string
	Terms    []string
	AllWords bool
}

// TemplatesFunc adds functions to templates
func TemplatesFunc(prod bool) *template.Template {
	return template.New("").Funcs(template.FuncMap{
		"assetUrl": func(name string, ext string) string {
			sep := "."
			if prod {
				sep = ".min."
			}
			return "/" + Version + "/" + ext + "/" + name + sep + ext
		},
		"humanSize": func(sz int64) string {
			return datasize.ByteSize(sz).HumanReadable()
		},
		"bookCover": func(book *BookFull) string {
			fmt.Println(book.Path)
			return "/calibre/" + url.PathEscape(book.Path) + "/cover.jpg"
		},
		"bookLink": func(data *BookData, book *BookFull) string {
			return "/calibre/" + url.PathEscape(book.Path) + "/" + url.PathEscape(data.Name) + "." + strings.ToLower(data.Format)
		},
	})
}

// RedirectHome redirects to home page
func RedirectHome(res http.ResponseWriter, req *http.Request) error {
	http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
	return nil
}

// output page with template
func (app *Bouquins) render(res http.ResponseWriter, tpl string, model interface{}) error {
	return app.Tpl.ExecuteTemplate(res, tpl, model)
}

// output as JSON
func writeJSON(res http.ResponseWriter, model interface{}) error {
	res.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(res)
	return enc.Encode(model)
}

// test if JSON requested
func isJSON(req *http.Request) bool {
	return req.Header.Get("Accept") == "application/json"
}

// get integer parameter
func paramInt(name string, req *http.Request) int {
	val := req.URL.Query().Get(name)
	if val == "" {
		return 0
	}
	valInt, err := strconv.Atoi(val)
	if err != nil {
		log.Println("Invalid  value for", name, ":", val)
		return 0
	}
	return valInt
}

// get order parameter
func paramOrder(req *http.Request) string {
	val := req.URL.Query().Get(pOrder)
	if val == "desc" || val == "asc" {
		return val
	}
	return ""
}

// get common request parameters
func params(req *http.Request) *ReqParams {
	page, perpage := paramInt(pPage, req), paramInt(pPerPage, req)
	limit := perpage
	if perpage == 0 {
		limit = defaultLimit
	}
	offset := perpage * (page - 1)
	if offset < 0 {
		offset = 0
	}
	sort := req.URL.Query().Get(pSort)
	order := paramOrder(req)
	terms := req.URL.Query()[pTerm]
	return &ReqParams{limit, offset, sort, order, terms, false}
}

// single element or list elements page
func listOrID(res http.ResponseWriter, req *http.Request, url string,
	listFunc func(res http.ResponseWriter, req *http.Request) error,
	idFunc func(idParam string, res http.ResponseWriter, req *http.Request) error) error {
	if !strings.HasPrefix(req.URL.Path, url) {
		return errors.New("Invalid URL") // FIXME 404
	}
	idParam := req.URL.Path[len(url):]
	if len(idParam) == 0 {
		return listFunc(res, req)
	}
	return idFunc(idParam, res, req)
}

// LIST ELEMENTS PAGES //

func (app *Bouquins) booksListPage(res http.ResponseWriter, req *http.Request) error {
	if isJSON(req) {
		books, count, more, err := app.BooksAdv(params(req))
		if err != nil {
			return err
		}
		return writeJSON(res, NewBooksResultsModel(books, more, count))
	}
	return errors.New("Invalid mime")
}
func (app *Bouquins) authorsListPage(res http.ResponseWriter, req *http.Request) error {
	if isJSON(req) {
		authors, count, more, err := app.AuthorsAdv(params(req))
		if err != nil {
			return err
		}
		return writeJSON(res, NewAuthorsResultsModel(authors, more, count))
	}
	return errors.New("Invalid mime")
}
func (app *Bouquins) seriesListPage(res http.ResponseWriter, req *http.Request) error {
	if isJSON(req) {
		series, count, more, err := app.SeriesAdv(params(req))
		if err != nil {
			return err
		}
		return writeJSON(res, NewSeriesResultsModel(series, more, count))
	}
	return errors.New("Invalid mime")
}

// SINGLE ELEMENT PAGES //

func (app *Bouquins) bookPage(idParam string, res http.ResponseWriter, req *http.Request) error {
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return err
	}
	book, err := app.BookFull(int64(id))
	if err != nil {
		return err
	}
	return app.render(res, tplBooks, &BookModel{*app.NewModel(book.Title, "book", req), book})
}
func (app *Bouquins) authorPage(idParam string, res http.ResponseWriter, req *http.Request) error {
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return err
	}
	author, err := app.AuthorFull(int64(id))
	if err != nil {
		return err
	}
	return app.render(res, tplAuthors, &AuthorModel{*app.NewModel(author.Name, "author", req), author})
}
func (app *Bouquins) seriePage(idParam string, res http.ResponseWriter, req *http.Request) error {
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return err
	}
	series, err := app.SeriesFull(int64(id))
	if err != nil {
		return err
	}
	return app.render(res, tplSeries, &SeriesModel{*app.NewModel(series.Name, "series", req), series})
}

// ROUTES //

// BooksPage displays a single books or a returns a list of books
func (app *Bouquins) BooksPage(res http.ResponseWriter, req *http.Request) error {
	return listOrID(res, req, URLBooks, app.booksListPage, app.bookPage)
}

// AuthorsPage displays a single author or returns a list of authors
func (app *Bouquins) AuthorsPage(res http.ResponseWriter, req *http.Request) error {
	return listOrID(res, req, URLAuthors, app.authorsListPage, app.authorPage)
}

// SeriesPage displays a single series or returns a list of series
func (app *Bouquins) SeriesPage(res http.ResponseWriter, req *http.Request) error {
	return listOrID(res, req, URLSeries, app.seriesListPage, app.seriePage)
}

// SearchPage displays search form and results
func (app *Bouquins) SearchPage(res http.ResponseWriter, req *http.Request) error {
	return app.render(res, tplSearch, app.NewSearchModel(req))
}

// AboutPage displays about page
func (app *Bouquins) AboutPage(res http.ResponseWriter, req *http.Request) error {
	return app.render(res, tplAbout, app.NewModel("A propos", "about", req))
}

// IndexPage displays index page: list of books/authors/series
func (app *Bouquins) IndexPage(res http.ResponseWriter, req *http.Request) error {
	count, err := app.BookCount()
	if err != nil {
		return err
	}
	model := app.NewIndexModel("", count, req)
	if isJSON(req) {
		return writeJSON(res, model)
	}
	return app.render(res, tplIndex, model)
}

func (app *Bouquins) CalibreFileServer() http.Handler {
	calibre := app.Conf.CalibrePath
	handler := http.StripPrefix(URLCalibre, http.FileServer(http.Dir(calibre)))
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		for _, suffix := range UnprotectedCalibreSuffix {
			if strings.HasSuffix(req.URL.Path, suffix) {
				handler.ServeHTTP(res, req)
			}
		}
		// check auth
		if app.Username(req) == "" {
			http.Error(res, "401 Unauthorized", http.StatusUnauthorized)
		} else {
			handler.ServeHTTP(res, req)
		}
	})
}
