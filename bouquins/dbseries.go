package bouquins

// MERGE SUB QUERIES //

func assignAuthorsSeries(series []*SeriesAdv, authors map[int64][]*Author) {
	for _, s := range series {
		s.Authors = authors[s.ID]
	}
}

// SUB QUERIES //

func (app *Bouquins) searchSeries(limit int, terms []string, all bool) ([]*SeriesAdv, int, error) {
	rows, err := app.searchHelper(all, terms, sqlSeriesSearch, sqlSeriesTerm, sqlSeriesOrder)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	series := make([]*SeriesAdv, 0, limit)
	count := 0
	for rows.Next() {
		if len(series) <= limit {
			serie := new(SeriesAdv)
			if err := rows.Scan(&serie.ID, &serie.Name); err != nil {
				return nil, 0, err
			}
			series = append(series, serie)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return series, count, nil
}

func (app *Bouquins) querySeriesList(limit, offset int, sort, order string) ([]*SeriesAdv, bool, error) {
	series := make([]*SeriesAdv, 0, limit)
	stmt, err := app.psSortSeries(qtSeries, sort, order)
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
		if len(series) == limit {
			more = true
		} else {
			serie := new(SeriesAdv)
			if err := rows.Scan(&serie.ID, &serie.Name, &serie.Count); err != nil {
				return nil, false, err
			}
			series = append(series, serie)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	return series, more, nil
}
func (app *Bouquins) querySeriesListAuthors(limit, offset int, sort, order string) (map[int64][]*Author, error) {
	authors := make(map[int64][]*Author)
	stmt, err := app.psSortBooks(qtSeriesAuthors, sort, order)
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
		var seriesID int64
		if err := rows.Scan(&author.ID, &author.Name, &seriesID); err != nil {
			return nil, err
		}
		if authors[seriesID] == nil {
			authors[seriesID] = append(make([]*Author, 0), author)
		} else {
			authors[seriesID] = append(authors[seriesID], author)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return authors, nil
}

func (app *Bouquins) querySeries(id int64) (*SeriesFull, error) {
	stmt, err := app.ps(qtSerie)
	if err != nil {
		return nil, err
	}
	series := new(SeriesFull)
	err = stmt.QueryRow(id).Scan(&series.ID, &series.Name)
	return series, nil
}
func (app *Bouquins) querySeriesAuthors(series *SeriesFull) error {
	stmt, err := app.ps(qtSerieAuthors)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(series.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		author := new(Author)
		if err = rows.Scan(&author.ID, &author.Name); err != nil {
			return err
		}
		series.Authors = append(series.Authors, author)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
func (app *Bouquins) querySeriesBooks(series *SeriesFull) error {
	stmt, err := app.ps(qtSerieBooks)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(series.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		book := new(Book)
		if err = rows.Scan(&book.ID, &book.Title, &book.SeriesIndex); err != nil {
			return err
		}
		series.Books = append(series.Books, book)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

// DB LOADS //

// SeriesFull loads a series
func (app *Bouquins) SeriesFull(id int64) (*SeriesFull, error) {
	series, err := app.querySeries(id)
	if err != nil {
		return nil, err
	}
	err = app.querySeriesBooks(series)
	if err != nil {
		return nil, err
	}
	err = app.querySeriesAuthors(series)
	if err != nil {
		return nil, err
	}
	return series, nil
}

// SeriesAdv loads a list of series
func (app *Bouquins) SeriesAdv(params *ReqParams) ([]*SeriesAdv, int, bool, error) {
	limit, offset, sort, order := params.Limit, params.Offset, params.Sort, params.Order
	if len(params.Terms) > 0 {
		series, count, err := app.searchSeries(limit, params.Terms, params.AllWords)
		return series, count, count > limit, err
	}
	series, more, err := app.querySeriesList(limit, offset, sort, order)
	if err != nil {
		return nil, 0, false, err
	}
	authors, err := app.querySeriesListAuthors(limit, offset, sort, order)
	if err != nil {
		return nil, 0, false, err
	}
	assignAuthorsSeries(series, authors)
	return series, 0, more, nil
}
