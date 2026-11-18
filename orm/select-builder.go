package orm

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type SelectQBuilder struct {
	db          *sql.DB
	debug       bool
	dialect     Dialect
	count       bool
	distinct    bool
	columnNames []string
	fields      []string
	from        []string
	where       string
	order       [][2]string
	limit       int64
	offset      int64
	binds       []interface{}
	err         bool
	errStr      string
}

// Crea e ritorna un nuovo SelectQBuilder
func SelectQuery(db *sql.DB, debug bool) *SelectQBuilder {
	return &SelectQBuilder{
		db:          db,
		debug:       debug,
		dialect:     DefaultDialect,
		columnNames: []string{},
		fields:      []string{},
		from:        []string{},
		order:       [][2]string{},
		limit:       -1,
		offset:      -1,
		binds:       []interface{}{},
	}
}

func (q *SelectQBuilder) fail(msg string) *SelectQBuilder {
	q.err = true
	q.errStr = msg
	return q
}

// Dialect sceglie la sintassi SQL generata (default: DefaultDialect).
func (q *SelectQBuilder) Dialect(d Dialect) *SelectQBuilder {
	if !d.valid() {
		return q.fail("invalid dialect")
	}
	q.dialect = d
	return q
}

// COUNT
func (q *SelectQBuilder) Count() *SelectQBuilder {
	q.count = true
	return q
}

// DISTINCT
func (q *SelectQBuilder) Distinct() *SelectQBuilder {
	q.distinct = true
	return q
}

// SELECT. Formati ammessi: "ID", "ID AS id", "c.ID", "c.ID AS id".
func (q *SelectQBuilder) Select(fields ...string) *SelectQBuilder {
	for _, f := range fields {
		switch {
		case identRe.MatchString(f):
			q.fields = append(q.fields, fmt.Sprintf("`%s`", f))
			q.columnNames = append(q.columnNames, f)
		case identAliasRe.MatchString(f):
			c := strings.Split(f, " ") // c[0] Colonna. c[1] AS. c[2] Alias.
			q.fields = append(q.fields, fmt.Sprintf("`%s` AS `%s`", c[0], c[2]))
			q.columnNames = append(q.columnNames, c[2])
		case qualIdentRe.MatchString(f):
			_, col, _ := strings.Cut(f, ".")
			q.fields = append(q.fields, quoteQualified(f))
			q.columnNames = append(q.columnNames, col)
		case qualIdentAliasRe.MatchString(f):
			c := strings.Split(f, " ") // c[0] Tabella.Colonna. c[1] AS. c[2] Alias.
			q.fields = append(q.fields, fmt.Sprintf("%s AS `%s`", quoteQualified(c[0]), c[2]))
			q.columnNames = append(q.columnNames, c[2])
		default:
			q.fields = []string{}
			q.columnNames = []string{}
			return q.fail("invalid SELECT field")
		}
	}
	return q
}

// FROM. Formati ammessi: "clienti", "clienti AS c".
func (q *SelectQBuilder) From(inputs ...string) *SelectQBuilder {
	for _, input := range inputs {
		switch {
		case identRe.MatchString(input):
			q.from = append(q.from, fmt.Sprintf("`%s`", input))
		case identAliasRe.MatchString(input):
			c := strings.Split(input, " ") // c[0] Nome tabella. c[1] AS. c[2] alias.
			q.from = append(q.from, fmt.Sprintf("`%s` AS `%s`", c[0], c[2]))
		default:
			q.from = []string{}
			return q.fail("invalid FROM table")
		}
	}
	return q
}

// WHERE (vedi parseWhere per la sintassi; ammette anche tabella.colonna).
func (q *SelectQBuilder) Where(inputs ...interface{}) *SelectQBuilder {
	where, binds, err := parseWhere(inputs, true)
	if err != nil {
		q.where = ""
		q.binds = []interface{}{}
		return q.fail(err.Error())
	}
	q.where += where
	q.binds = append(q.binds, binds...)
	return q
}

// ORDER. Ogni elemento è {colonna, direzione}; la direzione può essere
// solo "ASC" o "DESC" (maiuscole/minuscole indifferenti, vuota = ASC).
func (q *SelectQBuilder) Order(order [][2]string) *SelectQBuilder {
	// Lavora su una copia per non modificare la slice del chiamante.
	out := make([][2]string, 0, len(order))
	for _, o := range order {
		var col string
		switch {
		case identRe.MatchString(o[0]):
			col = fmt.Sprintf("`%s`", o[0])
		case qualIdentRe.MatchString(o[0]):
			col = quoteQualified(o[0])
		default:
			q.order = [][2]string{}
			return q.fail("invalid ORDER BY column")
		}

		// La direzione finisce nella query così com'è: accetta solo ASC/DESC.
		var dir string
		switch strings.ToUpper(strings.TrimSpace(o[1])) {
		case "", "ASC":
			dir = "ASC"
		case "DESC":
			dir = "DESC"
		default:
			q.order = [][2]string{}
			return q.fail("invalid ORDER BY direction")
		}
		out = append(out, [2]string{col, dir})
	}

	q.order = out
	return q
}

// LIMIT
func (q *SelectQBuilder) Limit(limit int64) *SelectQBuilder {
	q.limit = limit
	return q
}

// OFFSET
func (q *SelectQBuilder) Offset(offset int64) *SelectQBuilder {
	q.offset = offset
	return q
}

// toSQL costruisce la query, i binds e i nomi delle colonne del risultato,
// senza eseguirla.
func (q *SelectQBuilder) toSQL() (string, []interface{}, []string, error) {
	if q.err {
		return "", nil, nil, fmt.Errorf("%s", q.errStr)
	}
	if len(q.from) == 0 {
		return "", nil, nil, fmt.Errorf("missing FROM")
	}

	var s strings.Builder
	columnNames := q.columnNames
	if q.count {
		s.WriteString("SELECT COUNT(*) AS `CNT`")
		columnNames = []string{"CNT"}
	} else {
		if len(q.fields) == 0 {
			return "", nil, nil, fmt.Errorf("missing SELECT fields")
		}
		if q.distinct {
			s.WriteString("SELECT DISTINCT ")
		} else {
			s.WriteString("SELECT ")
		}
		s.WriteString(strings.Join(q.fields, ", "))
	}

	s.WriteString(" FROM ")
	s.WriteString(strings.Join(q.from, ", "))

	if q.where != "" {
		s.WriteString(" WHERE ")
		s.WriteString(q.where)
	}
	if len(q.order) > 0 {
		parts := make([]string, len(q.order))
		for i, ord := range q.order {
			parts[i] = ord[0] + " " + ord[1]
		}
		s.WriteString(" ORDER BY ")
		s.WriteString(strings.Join(parts, ", "))
	}
	if q.limit > 0 {
		// "LIMIT n OFFSET m" è supportata da MySQL, MariaDB, PostgreSQL e SQLite.
		s.WriteString(" LIMIT ")
		s.WriteString(strconv.FormatInt(q.limit, 10))
		if q.offset > 0 {
			s.WriteString(" OFFSET ")
			s.WriteString(strconv.FormatInt(q.offset, 10))
		}
	}
	n := 0
	return q.dialect.render(s.String(), &n), q.binds, columnNames, nil
}

// Build costruisce la query finale e la esegue.
func (q *SelectQBuilder) Build() Res {
	query, binds, columnNames, err := q.toSQL()
	if err != nil {
		return Res{Err: true, Msg: "Error while building the query: " + err.Error() + ".", Data: []map[string]interface{}{}}
	}

	// Esegui query
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}

	rows, err := q.db.Query(query, binds...)
	if err != nil {
		return *ManageErr(&res, q.debug, err, query)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("[DEBUG] Error deferring rows.Close():", err)
		}
	}()

	record := make([]interface{}, len(columnNames))
	for i := range record {
		record[i] = new(sql.RawBytes)
	}

	rv := []map[string]interface{}{}
	for rows.Next() {
		if err := rows.Scan(record...); err != nil {
			return *ManageErr(&res, q.debug, err, query)
		}
		vals := map[string]interface{}{}
		for i, k := range columnNames {
			// sql.RawBytes è un alias per []byte
			raw := *record[i].(*sql.RawBytes)
			if raw == nil {
				vals[k] = nil
			} else {
				vals[k] = string(raw)
			}
		}
		rv = append(rv, vals)
	}
	// Errori durante l'iterazione (es. connessione caduta a metà).
	if err := rows.Err(); err != nil {
		return *ManageErr(&res, q.debug, err, query)
	}

	res.Data = rv
	return res
}

/* Example
res := SelectQuery(db, false).
	Select("ID", "c.name", "email").
	From("clienti AS c").
	Where("((", "id = ?", 1, ")", "AND", "(", "c.name IS NULL", "))", "OR", "surname = ?", "Pluto").
	Order([][2]string{{"ID", "ASC"}, {"name", "DESC"}}).
	Limit(10).Offset(0).
	Build()
*/
