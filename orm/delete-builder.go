package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

type DeleteQBuilder struct {
	db      *sql.DB
	debug   bool
	dialect Dialect
	delete  string
	where   string
	binds   []interface{}
	allRows bool
	err     bool
	errStr  string
}

// Crea e ritorna un nuovo DeleteQBuilder
func DeleteQuery(db *sql.DB, debug bool) *DeleteQBuilder {
	return &DeleteQBuilder{
		db:      db,
		debug:   debug,
		dialect: DefaultDialect,
		delete:  "",
		where:   "",
		binds:   []interface{}{},
		err:     false,
	}
}

func (q *DeleteQBuilder) fail(msg string) *DeleteQBuilder {
	q.err = true
	q.errStr = msg
	return q
}

// Dialect sceglie la sintassi SQL generata (default: DefaultDialect).
func (q *DeleteQBuilder) Dialect(d Dialect) *DeleteQBuilder {
	if !d.valid() {
		return q.fail("invalid dialect")
	}
	q.dialect = d
	return q
}

// DELETE
func (q *DeleteQBuilder) Delete(input string) *DeleteQBuilder {
	if !identRe.MatchString(input) {
		q.delete = ""
		return q.fail("invalid table name")
	}
	q.delete = fmt.Sprintf("`%s`", input)
	return q
}

// WHERE (vedi parseWhere per la sintassi).
func (q *DeleteQBuilder) Where(inputs ...interface{}) *DeleteQBuilder {
	where, binds, err := parseWhere(inputs, false)
	if err != nil {
		q.where = ""
		q.binds = []interface{}{}
		return q.fail(err.Error())
	}
	q.where += where
	q.binds = append(q.binds, binds...)
	return q
}

// AllRows permette esplicitamente una DELETE senza WHERE (cancella tutta la tabella).
// Senza questa chiamata, Build() rifiuta una DELETE senza condizioni.
func (q *DeleteQBuilder) AllRows() *DeleteQBuilder {
	q.allRows = true
	return q
}

// toSQL costruisce la query e i binds senza eseguirla.
func (q *DeleteQBuilder) toSQL() (string, []interface{}, error) {
	if q.err {
		return "", nil, fmt.Errorf("%s", q.errStr)
	}
	if q.delete == "" {
		return "", nil, fmt.Errorf("missing table name")
	}
	if q.where == "" && !q.allRows {
		return "", nil, fmt.Errorf("DELETE without WHERE: call AllRows() to delete every row")
	}
	var s strings.Builder
	s.WriteString("DELETE FROM ")
	s.WriteString(q.delete)
	if q.where != "" {
		s.WriteString(" WHERE ")
		s.WriteString(q.where)
	}
	n := 0
	return q.dialect.render(s.String(), &n), q.binds, nil
}

// Build costruisce la query finale e la esegue.
func (q *DeleteQBuilder) Build() Res {
	query, binds, err := q.toSQL()
	if err != nil {
		return Res{Err: true, Msg: "Error while building the query: " + err.Error() + ".", Data: []map[string]interface{}{}}
	}

	// Esegui query
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}
	row, err := q.db.Exec(query, binds...)
	if err != nil {
		return *ManageErr(&res, q.debug, err, query)
	}
	rowsAffected, err := row.RowsAffected()
	if err != nil {
		rowsAffected = int64(0)
	}
	res.Data = append(res.Data, map[string]interface{}{"rowsAffected": rowsAffected})
	return res
}

/* Example

DELETE FROM Customers
WHERE CustomerID = 1;

	res := DeleteQuery(db, false).
		Delete("clienti").
		Where("ID = ?", 1).
		Build()
*/
