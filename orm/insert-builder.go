package orm

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

var (
	insTableNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

type InsertQBuilder struct {
	db        *sql.DB
	debug     bool
	insert    string
	columns   []string
	valuesStr string
	binds     []interface{}
	err       bool
	errStr    string
}

// Crea e ritorna un nuovo InsertQBuilder
func InsertQuery(db *sql.DB, debug bool) *InsertQBuilder {
	return &InsertQBuilder{
		db:        db,
		debug:     debug,
		insert:    "",
		columns:   []string{},
		valuesStr: "",
		binds:     []interface{}{},
		err:       false,
		errStr:    "",
	}
}

// INSERT
func (q *InsertQBuilder) Insert(input string) *InsertQBuilder {
	if insTableNameRegex.Match([]byte(input)) {
		q.insert = fmt.Sprintf("`%s`", input)
	} else {
		// Errore.
		q.err = true
		q.errStr = "table name regexp error"
		q.insert = ""
	}
	return q
}

// COLUMNS
func (q *InsertQBuilder) Columns(inputs ...string) *InsertQBuilder {
	for _, input := range inputs {
		if insTableNameRegex.Match([]byte(input)) {
			q.columns = append(q.columns, input)
		} else {
			// Errore.
			q.err = true
			q.errStr = "table name regexp error"
			q.columns = []string{}
			break
		}
	}
	return q
}

// VALUES
func (q *InsertQBuilder) Values(inputs ...interface{}) *InsertQBuilder {
	for i, input := range inputs {
		q.binds = append(q.binds, input)
		if i < len(inputs)-1 {
			q.valuesStr += "?, "
		} else {
			q.valuesStr += "?"
		}
	}
	return q
}

// Build costruisce la query finale e la esegue.
func (q *InsertQBuilder) Build() Res {
	if q.err {
		return Res{Err: true, Msg: "Error while building the query: " + q.errStr + ".", Data: []map[string]interface{}{}}
	}
	var s strings.Builder
	if q.insert != "" {
		s.WriteString("INSERT INTO ")
		s.WriteString(q.insert)
		s.WriteString("(")
		for i, k := range q.columns {
			if i < len(q.columns)-1 {
				s.WriteString(fmt.Sprintf("`%s`, ", k))
			} else {
				s.WriteString(fmt.Sprintf("`%s`", k))
			}
		}
		s.WriteString(")")
		s.WriteString(" VALUES(")
		s.WriteString(q.valuesStr)
		s.WriteString(")")
	}

	// Esegui query
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}
	row, err := q.db.Exec(s.String(), q.binds...)
	if err != nil {
		return *ManageErr(&res, q.debug, err, s.String())
	} else {
		var lastInsertId int64
		lastInsertId, _ = row.LastInsertId()
		res.Err = false
		res.Msg = ""
		res.Data = append(res.Data, map[string]interface{}{"lastInsertId": lastInsertId})
		return res
	}
}

/* Example
fields := []string{"email"}
query, binds := Query().
	Insert("clienti").Columns("name", "email").Values("Andrea", nil)
	Build()

fmt.Println(query)
fmt.Println(binds)
*/
