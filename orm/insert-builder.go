package orm

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type InsertQBuilder struct {
	db        *sql.DB
	debug     bool
	dialect   Dialect
	insert    string
	columns   []string
	binds     []interface{}
	returning string
	err       bool
	errStr    string
}

// Crea e ritorna un nuovo InsertQBuilder
func InsertQuery(db *sql.DB, debug bool) *InsertQBuilder {
	return &InsertQBuilder{
		db:      db,
		debug:   debug,
		dialect: DefaultDialect,
		columns: []string{},
		binds:   []interface{}{},
	}
}

func (q *InsertQBuilder) fail(msg string) *InsertQBuilder {
	q.err = true
	q.errStr = msg
	return q
}

// Dialect sceglie la sintassi SQL generata (default: DefaultDialect).
func (q *InsertQBuilder) Dialect(d Dialect) *InsertQBuilder {
	if !d.valid() {
		return q.fail("invalid dialect")
	}
	q.dialect = d
	return q
}

// RETURNING (solo Postgres): restituisce il valore della colonna indicata
// (tipicamente la chiave primaria generata) in Data[0]["lastInsertId"].
// I driver PostgreSQL non supportano LastInsertId(): senza Returning(),
// con Postgres lastInsertId vale sempre 0.
func (q *InsertQBuilder) Returning(column string) *InsertQBuilder {
	if !identRe.MatchString(column) {
		return q.fail("invalid RETURNING column")
	}
	q.returning = column
	return q
}

// INSERT
func (q *InsertQBuilder) Insert(input string) *InsertQBuilder {
	if !identRe.MatchString(input) {
		q.insert = ""
		return q.fail("table name regexp error")
	}
	q.insert = fmt.Sprintf("`%s`", input)
	return q
}

// COLUMNS
func (q *InsertQBuilder) Columns(inputs ...string) *InsertQBuilder {
	for _, input := range inputs {
		if !identRe.MatchString(input) {
			q.columns = []string{}
			return q.fail("column name regexp error")
		}
		q.columns = append(q.columns, input)
	}
	return q
}

// VALUES
func (q *InsertQBuilder) Values(inputs ...interface{}) *InsertQBuilder {
	q.binds = append(q.binds, inputs...)
	return q
}

// toSQL costruisce la query e i binds senza eseguirla.
func (q *InsertQBuilder) toSQL() (string, []interface{}, error) {
	if q.err {
		return "", nil, fmt.Errorf("%s", q.errStr)
	}
	if q.insert == "" {
		return "", nil, fmt.Errorf("missing table name")
	}
	if len(q.columns) == 0 {
		return "", nil, fmt.Errorf("missing columns")
	}
	if len(q.binds) != len(q.columns) {
		return "", nil, fmt.Errorf("%d columns but %d values", len(q.columns), len(q.binds))
	}

	cols := make([]string, len(q.columns))
	for i, c := range q.columns {
		cols[i] = fmt.Sprintf("`%s`", c)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(q.binds)), ", ")

	query := "INSERT INTO " + q.insert + "(" + strings.Join(cols, ", ") + ") VALUES(" + placeholders + ")"
	if q.returning != "" {
		if q.dialect != Postgres {
			return "", nil, fmt.Errorf("RETURNING is not supported by the %s dialect", q.dialect)
		}
		query += " RETURNING `" + q.returning + "`"
	}
	n := 0
	return q.dialect.render(query, &n), q.binds, nil
}

// Build costruisce la query finale e la esegue.
func (q *InsertQBuilder) Build() Res {
	query, binds, err := q.toSQL()
	if err != nil {
		return Res{Err: true, Msg: "Error while building the query: " + err.Error() + ".", Data: []map[string]interface{}{}}
	}

	// Esegui query
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}

	if q.returning != "" {
		var id sql.NullString
		if err := q.db.QueryRow(query, binds...).Scan(&id); err != nil {
			return *ManageErr(&res, q.debug, err, query)
		}
		res.Data = append(res.Data, map[string]interface{}{"lastInsertId": returnedID(id)})
		return res
	}

	row, err := q.db.Exec(query, binds...)
	if err != nil {
		return *ManageErr(&res, q.debug, err, query)
	}
	lastInsertId, _ := row.LastInsertId()
	res.Data = append(res.Data, map[string]interface{}{"lastInsertId": lastInsertId})
	return res
}

// returnedID converte il valore di RETURNING: int64 se è un intero
// (come LastInsertId), altrimenti la stringa (es. UUID), nil se NULL.
func returnedID(v sql.NullString) interface{} {
	if !v.Valid {
		return nil
	}
	if n, err := strconv.ParseInt(v.String, 10, 64); err == nil {
		return n
	}
	return v.String
}

/* Example
res := InsertQuery(db, false).
	Insert("clienti").
	Columns("name", "email").
	Values("Andrea", nil).
	Build()

Postgres:
res := InsertQuery(db, false).
	Dialect(Postgres).
	Insert("clienti").
	Columns("name", "email").
	Values("Andrea", nil).
	Returning("id").
	Build()
id := res.Data[0]["lastInsertId"]
*/
