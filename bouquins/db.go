package bouquins

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	sqlBooks0 = `SELECT books.id AS id,title,series_index,name as series_name,series.id AS series_id 
    FROM books LEFT OUTER JOIN books_series_link ON books.id = books_series_link.book 
    LEFT OUTER JOIN series ON series.id = books_series_link.series `
	sqlBooksTags0 = `SELECT name, books_tags_link.book as book FROM tags, books_tags_link 
    WHERE tags.id = books_tags_link.tag AND books_tags_link.book IN ( SELECT id FROM books `
	sqlBooksAuthors0 = `SELECT authors.id, authors.name, books_authors_link.book as book 
    FROM authors, books_authors_link WHERE books_authors_link.author = authors.id 
    AND books_authors_link.book IN ( SELECT id FROM books `
	sqlBooksTerm = " books.sort like ? "

	sqlSeries0 = `SELECT series.id, series.name, count(book) FROM series 
    LEFT OUTER JOIN books_series_link ON books_series_link.series = series.id 
    GROUP BY series.id `
	sqlSeriesAuthors0 = `SELECT DISTINCT authors.id, authors.name, books_series_link.series 
    FROM authors, books_authors_link, books_series_link 
    WHERE books_authors_link.book = books_series_link.book AND books_authors_link.author = authors.id 
    AND books_series_link.series IN ( SELECT id FROM series `
	sqlSeriesSearch = "SELECT series.id, series.name FROM series WHERE "
	sqlSeriesTerm   = " series.sort like ? "

	sqlAuthors0 = `SELECT authors.id, authors.name, count(book) as count FROM authors, books_authors_link 
    WHERE authors.id = books_authors_link.author GROUP BY author `
	sqlAuthorsSearch = "SELECT id, name FROM authors WHERE "
	sqlAuthorsTerm   = " sort like ? "

	sqlPage  = " LIMIT ? OFFSET ?"
	sqlWhere = " WHERE "
	sqlAnd   = " AND "
	sqlOr    = " OR "

	sqlBooksOrder   = " ORDER BY books.sort"
	sqlAuthorsOrder = " ORDER BY authors.sort"
	sqlSeriesOrder  = " ORDER BY series.sort"

	sqlBooksCount = "SELECT count(id) FROM books"
	sqlBook       = `SELECT books.id AS id,title, series_index, series.name AS series_name, series.id AS series_id, 
    strftime('%s', timestamp), strftime('%Y', pubdate), isbn,lccn,path,uuid,has_cover, 
    languages.lang_code, publishers.name AS pubname FROM books 
    LEFT OUTER JOIN books_languages_link ON books_languages_link.book = books.id 
    LEFT OUTER JOIN languages ON languages.id = books_languages_link.lang_code 
    LEFT OUTER JOIN data ON data.book = books.id 
    LEFT OUTER JOIN books_series_link ON books.id = books_series_link.book 
    LEFT OUTER JOIN series ON series.id = books_series_link.series 
    LEFT OUTER JOIN books_publishers_link ON books.id = books_publishers_link.book 
    LEFT OUTER JOIN publishers ON publishers.id = books_publishers_link.publisher 
    WHERE books.id = ?`
	sqlBookTags    = "SELECT name FROM tags, books_tags_link WHERE tags.id = books_tags_link.tag AND books_tags_link.book = ?"
	sqlBookAuthors = `SELECT authors.id, authors.name, books_authors_link.book as book 
    FROM authors, books_authors_link WHERE books_authors_link.author = authors.id 
    AND books_authors_link.book = ?`
	sqlBookData              = "SELECT data.name, data.format, data.uncompressed_size FROM data WHERE data.book = ?"
	sqlBooksIDAsc            = sqlBooks0 + " ORDER BY id" + sqlPage
	sqlBooksIDDesc           = sqlBooks0 + "ORDER BY id DESC" + sqlPage
	sqlBooksTitleAsc         = sqlBooks0 + "ORDER BY books.sort" + sqlPage
	sqlBooksTitleDesc        = sqlBooks0 + "ORDER BY books.sort DESC" + sqlPage
	sqlBooksTagsIDAsc        = sqlBooksTags0 + "ORDER BY id" + sqlPage + ")"
	sqlBooksTagsIDDesc       = sqlBooksTags0 + "ORDER BY id DESC" + sqlPage + ")"
	sqlBooksTagsTitleAsc     = sqlBooksTags0 + "ORDER BY books.sort" + sqlPage + ")"
	sqlBooksTagsTitleDesc    = sqlBooksTags0 + "ORDER BY books.sort DESC" + sqlPage + ")"
	sqlBooksAuthorsIDAsc     = sqlBooksAuthors0 + "ORDER BY id" + sqlPage + ")"
	sqlBooksAuthorsIDDesc    = sqlBooksAuthors0 + "ORDER BY id DESC" + sqlPage + ")"
	sqlBooksAuthorsTitleAsc  = sqlBooksAuthors0 + "ORDER BY books.sort" + sqlPage + ")"
	sqlBooksAuthorsTitleDesc = sqlBooksAuthors0 + "ORDER BY books.sort DESC" + sqlPage + ")"

	sqlSeriesIDAsc           = sqlSeries0 + " ORDER BY series.id" + sqlPage
	sqlSeriesIDDesc          = sqlSeries0 + " ORDER BY series.id DESC" + sqlPage
	sqlSeriesNameAsc         = sqlSeries0 + " ORDER BY series.sort" + sqlPage
	sqlSeriesNameDesc        = sqlSeries0 + " ORDER BY series.sort DESC" + sqlPage
	sqlSeriesAuthorsIDAsc    = sqlSeriesAuthors0 + " ORDER BY series.id" + sqlPage + ")"
	sqlSeriesAuthorsIDDesc   = sqlSeriesAuthors0 + " ORDER BY series.id DESC" + sqlPage + ")"
	sqlSeriesAuthorsNameAsc  = sqlSeriesAuthors0 + " ORDER BY series.sort" + sqlPage + ")"
	sqlSeriesAuthorsNameDesc = sqlSeriesAuthors0 + " ORDER BY series.sort DESC" + sqlPage + ")"
	sqlSerie                 = "SELECT series.id, series.name FROM series WHERE series.id = ?"
	sqlSerieBooks            = `SELECT books.id, title, series_index FROM books 
    LEFT OUTER JOIN books_series_link ON books.id = books_series_link.book 
    WHERE books_series_link.series = ? ORDER BY series_index ASC`
	sqlSerieAuthors = `SELECT DISTINCT authors.id, authors.name 
    FROM authors, books_authors_link, books_series_link 
    WHERE books_authors_link.book = books_series_link.book AND books_authors_link.author = authors.id 
    AND books_series_link.series = ?`

	sqlAuthorsIDAsc    = sqlAuthors0 + "ORDER BY authors.id " + sqlPage
	sqlAuthorsIDDesc   = sqlAuthors0 + "ORDER BY authors.id DESC " + sqlPage
	sqlAuthorsNameAsc  = sqlAuthors0 + "ORDER BY authors.sort " + sqlPage
	sqlAuthorsNameDesc = sqlAuthors0 + "ORDER BY authors.sort DESC " + sqlPage
	sqlAuthorBooks     = `SELECT books.id AS id,title,series_index,name as series_name,series.id AS series_id 
    FROM books LEFT OUTER JOIN books_series_link ON books.id = books_series_link.book 
    LEFT OUTER JOIN series ON series.id = books_series_link.series 
    LEFT OUTER JOIN books_authors_link ON books.id = books_authors_link.book 
    WHERE books_authors_link.author = ? ORDER BY id`
	sqlAuthorAuthors = `SELECT DISTINCT authors.id, authors.name
    FROM authors, books_authors_link WHERE books_authors_link.author = authors.id 
    AND books_authors_link.book IN ( SELECT books.id FROM books LEFT OUTER JOIN books_authors_link 
			ON books.id = books_authors_link.book WHERE books_authors_link.author = ? ORDER BY books.id)
		AND authors.id != ? ORDER BY authors.id`
	sqlAuthor = "SELECT name FROM authors WHERE id = ?"

	sqlAccount = "SELECT accounts.id, name FROM accounts, authentifiers WHERE authentifiers.id = accounts.id AND authentifiers.authentifier = ?"

	defaultLimit = 10

	qtBook QueryType = iota
	qtBookTags
	qtBookData
	qtBookAuthors
	qtBookCount
	qtBooks
	qtBooksTags
	qtBooksAuthors
	qtSerie
	qtSerieAuthors
	qtSerieBooks
	qtSeries
	qtSeriesAuthors
	qtAuthor
	qtAuthorBooks
	qtAuthorCoauthors
	qtAuthors
)

var queries = map[Query]string{
	Query{qtBooks, true, true}:             sqlBooksTitleDesc,
	Query{qtBooks, true, false}:            sqlBooksTitleAsc,
	Query{qtBooks, false, true}:            sqlBooksIDDesc,
	Query{qtBooks, false, false}:           sqlBooksIDAsc,
	Query{qtBooksTags, true, true}:         sqlBooksTagsTitleDesc,
	Query{qtBooksTags, true, false}:        sqlBooksTagsTitleAsc,
	Query{qtBooksTags, false, true}:        sqlBooksTagsIDDesc,
	Query{qtBooksTags, false, false}:       sqlBooksTagsIDAsc,
	Query{qtBooksAuthors, true, true}:      sqlBooksAuthorsTitleDesc,
	Query{qtBooksAuthors, true, false}:     sqlBooksAuthorsTitleAsc,
	Query{qtBooksAuthors, false, true}:     sqlBooksAuthorsIDDesc,
	Query{qtBooksAuthors, false, false}:    sqlBooksAuthorsIDAsc,
	Query{qtBook, false, false}:            sqlBook,
	Query{qtBookTags, false, false}:        sqlBookTags,
	Query{qtBookData, false, false}:        sqlBookData,
	Query{qtBookAuthors, false, false}:     sqlBookAuthors,
	Query{qtBookCount, false, false}:       sqlBooksCount,
	Query{qtSerie, false, false}:           sqlSerie,
	Query{qtSeries, true, true}:            sqlSeriesNameDesc,
	Query{qtSeries, true, false}:           sqlSeriesNameAsc,
	Query{qtSeries, false, true}:           sqlSeriesIDDesc,
	Query{qtSeries, false, false}:          sqlSeriesIDAsc,
	Query{qtSeriesAuthors, true, true}:     sqlSeriesAuthorsNameDesc,
	Query{qtSeriesAuthors, true, false}:    sqlSeriesAuthorsNameAsc,
	Query{qtSeriesAuthors, false, true}:    sqlSeriesAuthorsIDDesc,
	Query{qtSeriesAuthors, false, false}:   sqlSeriesAuthorsIDAsc,
	Query{qtSerieAuthors, false, false}:    sqlSerieAuthors,
	Query{qtSerieBooks, false, false}:      sqlSerieBooks,
	Query{qtAuthors, true, true}:           sqlAuthorsNameDesc,
	Query{qtAuthors, true, false}:          sqlAuthorsNameAsc,
	Query{qtAuthors, false, true}:          sqlAuthorsIDDesc,
	Query{qtAuthors, false, false}:         sqlAuthorsIDAsc,
	Query{qtAuthor, false, false}:          sqlAuthor,
	Query{qtAuthorBooks, false, false}:     sqlAuthorBooks,
	Query{qtAuthorCoauthors, false, false}: sqlAuthorAuthors,
}
var (
	stmts       = make(map[Query]*sql.Stmt)
	stmtAccount *sql.Stmt
)

// QueryType is a type of query, with variants for sort and order
type QueryType uint

// Query is a key for SQL queries catalog
type Query struct {
	Type      QueryType
	SortField bool // sort by name or title
	Desc      bool
}

func (app *Bouquins) searchHelper(all bool, terms []string, stub, termExpr, orderExpr string) (*sql.Rows, error) {
	query := stub
	queryTerms := make([]interface{}, 0, len(terms))
	for i, term := range terms {
		queryTerms = append(queryTerms, "%"+term+"%")
		query += termExpr
		if i < len(terms)-1 && all {
			query += sqlAnd
		}
		if i < len(terms)-1 && !all {
			query += sqlOr
		}
	}
	query += orderExpr
	log.Println("Search:", query)

	stmt, err := app.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(queryTerms...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// PREPARED STATEMENTS //

// PrepareAll prepares statement for (almost) all queries
func (app *Bouquins) PrepareAll() error {
	errcount := 0
	for q, v := range queries {
		stmt, err := app.DB.Prepare(v)
		if err != nil {
			log.Println(err, v)
			errcount++
		}
		stmts[q] = stmt
	}
	// users.db
	var err error
	stmtAccount, err = app.UserDB.Prepare(sqlAccount)
	if err != nil {
		log.Println(err, sqlAccount)
		errcount++
	}
	if errcount > 0 {
		return fmt.Errorf("%d errors on queries, see logs", errcount)
	}
	return nil
}

// prepared statement with sort on books
func (app *Bouquins) psSortBooks(qt QueryType, sort, order string) (*sql.Stmt, error) {
	return app.psSort("title", qt, sort, order)
}
func (app *Bouquins) psSortAuthors(qt QueryType, sort, order string) (*sql.Stmt, error) {
	return app.psSort("name", qt, sort, order)
}
func (app *Bouquins) psSortSeries(qt QueryType, sort, order string) (*sql.Stmt, error) {
	return app.psSort("name", qt, sort, order)
}
func (app *Bouquins) psSort(sortNameField string, qt QueryType, sort, order string) (*sql.Stmt, error) {
	q := Query{qt, sort == sortNameField, order == "desc"}
	query := queries[q]
	log.Println(query)
	stmt := stmts[q]
	if stmt == nil {
		log.Println("Missing statement for ", q)
		var err error
		stmt, err = app.DB.Prepare(query)
		if err != nil {
			return nil, err
		}
	}
	return stmt, nil
}

// prepared statement without sort
func (app *Bouquins) ps(qt QueryType) (*sql.Stmt, error) {
	return app.psSort("any", qt, "", "")
}
