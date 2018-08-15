package bouquins

import (
	"database/sql"
)

// MERGE SUB QUERIES //
func assignAuthorsTagsBooks(books []*BookAdv, authors map[int64][]*Author, tags map[int64][]string) {
	for _, b := range books {
		b.Authors = authors[b.ID]
		b.Tags = tags[b.ID]
	}
}

// SUB QUERIES //

func (app *Bouquins) searchBooks(limit int, terms []string, all bool) ([]*BookAdv, int, error) {
	rows, err := app.searchHelper(all, terms, sqlBooks0+sqlWhere, sqlBooksTerm, sqlBooksOrder)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	books := make([]*BookAdv, 0, limit)
	count := 0
	for rows.Next() {
		if len(books) <= limit {
			book := new(BookAdv)
			var seriesName sql.NullString
			var seriesID sql.NullInt64
			if err := rows.Scan(&book.ID, &book.Title, &book.SeriesIndex, &seriesName, &seriesID); err != nil {
				return nil, 0, err
			}
			if seriesName.Valid && seriesID.Valid {
				book.Series = &Series{
					seriesID.Int64,
					seriesName.String,
				}
			}
			books = append(books, book)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return books, count, nil
}

func (app *Bouquins) queryBooks(limit, offset int, sort, order string) ([]*BookAdv, bool, error) {
	books := make([]*BookAdv, 0, limit)
	stmt, err := app.psSortBooks(qtBooks, sort, order)
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
		if len(books) == limit {
			more = true
		} else {
			book := new(BookAdv)
			var seriesName sql.NullString
			var seriesID sql.NullInt64
			if err := rows.Scan(&book.ID, &book.Title, &book.SeriesIndex, &seriesName, &seriesID); err != nil {
				return nil, false, err
			}
			if seriesName.Valid && seriesID.Valid {
				book.Series = &Series{
					seriesID.Int64,
					seriesName.String,
				}
			}
			books = append(books, book)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	return books, more, nil
}

func (app *Bouquins) queryBooksAuthors(limit, offset int, sort, order string) (map[int64][]*Author, error) {
	authors := make(map[int64][]*Author)
	stmt, err := app.psSortBooks(qtBooksAuthors, sort, order)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		author := new(Author)
		var book int64
		if err := rows.Scan(&author.ID, &author.Name, &book); err != nil {
			return nil, err
		}
		if authors[book] == nil {
			authors[book] = append(make([]*Author, 0), author)
		} else {
			authors[book] = append(authors[book], author)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return authors, nil
}

func (app *Bouquins) queryBooksTags(limit, offset int, sort, order string) (map[int64][]string, error) {
	stmt, err := app.psSortBooks(qtBooksTags, sort, order)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := make(map[int64][]string)
	for rows.Next() {
		var tag string
		var book int64
		if err := rows.Scan(&tag, &book); err != nil {
			return nil, err
		}
		bookTags := tags[book]
		if bookTags == nil {
			bookTags = make([]string, 0)
			tags[book] = bookTags
		}
		tags[book] = append(bookTags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

func (app *Bouquins) queryBook(id int64) (*BookFull, error) {
	stmt, err := app.ps(qtBook)
	if err != nil {
		return nil, err
	}
	book := new(BookFull)
	var seriesIdx sql.NullFloat64
	var seriesID, timestamp, pubdate sql.NullInt64
	var seriesName, isbn, lccn, uuid, lang, publisher sql.NullString
	var cover sql.NullBool
	err = stmt.QueryRow(id).Scan(&book.ID, &book.Title, &seriesIdx, &seriesName, &seriesID,
		&timestamp, &pubdate, &isbn, &lccn, &book.Path, &uuid, &cover, &lang, &publisher)
	if err != nil {
		return nil, err
	}
	if seriesID.Valid && seriesName.Valid && seriesIdx.Valid {
		book.SeriesIndex = seriesIdx.Float64
		book.Series = &Series{seriesID.Int64, seriesName.String}
	}
	if timestamp.Valid {
		book.Timestamp = timestamp.Int64
	}
	if pubdate.Valid {
		book.Pubdate = pubdate.Int64
	}
	if isbn.Valid {
		book.Isbn = isbn.String
	}
	if lccn.Valid {
		book.Lccn = lccn.String
	}
	if uuid.Valid {
		book.UUID = uuid.String
	}
	if lang.Valid {
		book.Lang = lang.String
	}
	if publisher.Valid {
		book.Publisher = publisher.String
	}
	if cover.Valid {
		book.HasCover = cover.Bool
	}
	return book, nil
}
func (app *Bouquins) queryBookTags(book *BookFull) error {
	stmt, err := app.ps(qtBookTags)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(book.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return err
		}
		book.Tags = append(book.Tags, tag)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
func (app *Bouquins) queryBookData(book *BookFull) error {
	stmt, err := app.ps(qtBookData)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(book.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		data := new(BookData)
		if err = rows.Scan(&data.Name, &data.Format, &data.Size); err != nil {
			return err
		}
		book.Data = append(book.Data, data)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
func (app *Bouquins) queryBookAuthors(book *BookFull) error {
	stmt, err := app.ps(qtBookAuthors)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(book.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		author := new(Author)
		var bookID int64
		if err = rows.Scan(&author.ID, &author.Name, &bookID); err != nil {
			return err
		}
		book.Authors = append(book.Authors, author)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

// DB LOADS //

// BookCount counts books in database
func (app *Bouquins) BookCount() (int64, error) {
	var count int64
	stmt, err := app.ps(qtBookCount)
	if err != nil {
		return 0, err
	}
	row := stmt.QueryRow()
	err = row.Scan(&count)
	return count, err
}

// BookFull loads a book
func (app *Bouquins) BookFull(id int64) (*BookFull, error) {
	book, err := app.queryBook(id)
	if err != nil {
		return nil, err
	}
	err = app.queryBookTags(book)
	if err != nil {
		return nil, err
	}
	err = app.queryBookAuthors(book)
	if err != nil {
		return nil, err
	}
	err = app.queryBookData(book)
	if err != nil {
		return nil, err
	}
	return book, nil
}

// BooksAdv loads a list of books
func (app *Bouquins) BooksAdv(params *ReqParams) ([]*BookAdv, int, bool, error) {
	limit, offset, sort, order := params.Limit, params.Offset, params.Sort, params.Order
	if len(params.Terms) > 0 {
		books, count, err := app.searchBooks(limit, params.Terms, params.AllWords)
		return books, count, count > limit, err
	}
	books, more, err := app.queryBooks(limit, offset, sort, order)
	if err != nil {
		return nil, 0, false, err
	}
	authors, err := app.queryBooksAuthors(limit, offset, sort, order)
	if err != nil {
		return nil, 0, false, err
	}
	tags, err := app.queryBooksTags(limit, offset, sort, order)
	if err != nil {
		return nil, 0, false, err
	}
	assignAuthorsTagsBooks(books, authors, tags)
	return books, 0, more, nil
}
