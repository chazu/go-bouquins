package bouquins

import (
	"database/sql"
)

// SUB QUERIES //

func (app *Bouquins) searchAuthors(limit int, terms []string, all bool) ([]*AuthorAdv, int, error) {
	rows, err := app.searchHelper(all, terms, sqlAuthorsSearch, sqlAuthorsTerm, sqlAuthorsOrder)
	if err != nil {
		return nil, 0, err
	}
	authors := make([]*AuthorAdv, 0, limit)
	count := 0
	defer rows.Close()
	for rows.Next() {
		if len(authors) <= limit {
			author := new(AuthorAdv)
			if err := rows.Scan(&author.ID, &author.Name); err != nil {
				return nil, 0, err
			}
			authors = append(authors, author)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return authors, count, nil
}

func (app *Bouquins) queryAuthors(limit, offset int, sort, order string) ([]*AuthorAdv, bool, error) {
	authors := make([]*AuthorAdv, 0, limit)
	stmt, err := app.psSortAuthors(qtAuthors, sort, order)
	if err != nil {
		return nil, false, err
	}
	rows, err := stmt.Query(limit+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	more := false
	for rows.Next() {
		if len(authors) == limit {
			more = true
		} else {
			author := new(AuthorAdv)
			if err := rows.Scan(&author.ID, &author.Name, &author.Count); err != nil {
				return nil, false, err
			}
			authors = append(authors, author)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	return authors, more, nil
}

func (app *Bouquins) queryAuthorBooks(author *AuthorFull) error {
	stmt, err := app.ps(qtAuthorBooks)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(author.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	series := make(map[int64]*Series, 0)
	for rows.Next() {
		book := new(Book)
		var seriesID sql.NullInt64
		var seriesName sql.NullString
		if err = rows.Scan(&book.ID, &book.Title, &book.SeriesIndex, &seriesName, &seriesID); err != nil {
			return err
		}
		if seriesID.Valid && seriesName.Valid {
			series[seriesID.Int64] = &Series{
				seriesID.Int64,
				seriesName.String,
			}
		}
		author.Books = append(author.Books, book)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	author.Series = make([]*Series, 0, len(series))
	for _, s := range series {
		author.Series = append(author.Series, s)
	}
	return nil
}

func (app *Bouquins) queryAuthorAuthors(author *AuthorFull) error {
	stmt, err := app.ps(qtAuthorCoauthors)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(author.ID, author.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		coauthor := new(Author)
		if err = rows.Scan(&coauthor.ID, &coauthor.Name); err != nil {
			return err
		}
		author.CoAuthors = append(author.CoAuthors, coauthor)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func (app *Bouquins) queryAuthor(id int64) (*AuthorFull, error) {
	stmt, err := app.ps(qtAuthor)
	if err != nil {
		return nil, err
	}
	author := new(AuthorFull)
	author.ID = id
	err = stmt.QueryRow(id).Scan(&author.Name)
	if err != nil {
		return nil, err
	}
	return author, nil
}

// DB LOADS //

// AuthorsAdv loads a list of authors
func (app *Bouquins) AuthorsAdv(params *ReqParams) ([]*AuthorAdv, int, bool, error) {
	limit, offset, sort, order := params.Limit, params.Offset, params.Sort, params.Order
	if len(params.Terms) > 0 {
		authors, count, err := app.searchAuthors(limit, params.Terms, params.AllWords)
		return authors, count, count > limit, err
	}
	authors, more, err := app.queryAuthors(limit, offset, sort, order)
	if err != nil {
		return nil, 0, false, err
	}
	return authors, 0, more, nil
}

// AuthorFull loads an author
func (app *Bouquins) AuthorFull(id int64) (*AuthorFull, error) {
	author, err := app.queryAuthor(id)
	if err != nil {
		return nil, err
	}
	err = app.queryAuthorBooks(author)
	if err != nil {
		return nil, err
	}
	err = app.queryAuthorAuthors(author)
	if err != nil {
		return nil, err
	}
	return author, nil
}
