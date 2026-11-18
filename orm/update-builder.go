package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

type UpdateQBuilder struct {
	db         *sql.DB
	debug      bool
	dialect    Dialect
	update     string
	set        []string
	setRaw     []string
	setBinds   []interface{}
	where      string
	whereBinds []interface{}
	allRows    bool
	err        bool
	errStr     string
}

// Crea e ritorna un nuovo UpdateQBuilder
func UpdateQuery(db *sql.DB, debug bool) *UpdateQBuilder {
	return &UpdateQBuilder{
		db:         db,
		debug:      debug,
		dialect:    DefaultDialect,
		set:        []string{},
		setRaw:     []string{},
		setBinds:   []interface{}{},
		whereBinds: []interface{}{},
	}
}

func (q *UpdateQBuilder) fail(msg string) *UpdateQBuilder {
	q.err = true
	q.errStr = msg
	return q
}

// Dialect sceglie la sintassi SQL generata (default: DefaultDialect).
func (q *UpdateQBuilder) Dialect(d Dialect) *UpdateQBuilder {
	if !d.valid() {
		return q.fail("invalid dialect")
	}
	q.dialect = d
	return q
}

// UPDATE
func (q *UpdateQBuilder) Update(input string) *UpdateQBuilder {
	if !identRe.MatchString(input) {
		q.update = ""
		return q.fail("invalid table name")
	}
	q.update = fmt.Sprintf("`%s`", input)
	return q
}

// SET: coppie "colonna = ?", valore. Es. Set("name = ?", "Andrea", "age = ?", 30).
func (q *UpdateQBuilder) Set(inputs ...interface{}) *UpdateQBuilder {
	if len(inputs)%2 != 0 {
		return q.fail("SET expects pairs of \"column = ?\" and value")
	}
	for i := 0; i < len(inputs); i += 2 {
		cond, ok := inputs[i].(string)
		if !ok || !setCondRe.MatchString(cond) {
			q.set = []string{}
			q.setBinds = []interface{}{}
			return q.fail("invalid SET expression")
		}
		col, _, _ := strings.Cut(cond, " ")
		q.set = append(q.set, fmt.Sprintf("`%s` = ?", col))
		q.setBinds = append(q.setBinds, inputs[i+1])
	}
	return q
}

// SET_RAW aggiunge espressioni SET scritte a mano, es. SetRaw("`counter` = `counter` + 1").
//
// ATTENZIONE: il testo finisce nella query così com'è, senza validazione né escaping.
// Non passare MAI dati provenienti dall'utente: usa Set() con i placeholder.
// Il testo non viene convertito al Dialect: con Postgres scrivi "counter" = "counter" + 1.
func (q *UpdateQBuilder) SetRaw(inputs ...string) *UpdateQBuilder {
	q.setRaw = append(q.setRaw, inputs...)
	return q
}

// WHERE (vedi parseWhere per la sintassi).
func (q *UpdateQBuilder) Where(inputs ...interface{}) *UpdateQBuilder {
	where, binds, err := parseWhere(inputs, false)
	if err != nil {
		q.where = ""
		q.whereBinds = []interface{}{}
		return q.fail(err.Error())
	}
	q.where += where
	q.whereBinds = append(q.whereBinds, binds...)
	return q
}

// AllRows permette esplicitamente una UPDATE senza WHERE (modifica tutta la tabella).
// Senza questa chiamata, Build() rifiuta una UPDATE senza condizioni.
func (q *UpdateQBuilder) AllRows() *UpdateQBuilder {
	q.allRows = true
	return q
}

// toSQL costruisce la query e i binds senza eseguirla.
// I binds del SET precedono sempre quelli del WHERE, in qualunque ordine
// siano stati chiamati Set() e Where().
func (q *UpdateQBuilder) toSQL() (string, []interface{}, error) {
	if q.err {
		return "", nil, fmt.Errorf("%s", q.errStr)
	}
	if q.update == "" {
		return "", nil, fmt.Errorf("missing table name")
	}
	if len(q.set)+len(q.setRaw) == 0 {
		return "", nil, fmt.Errorf("missing SET")
	}
	if q.where == "" && !q.allRows {
		return "", nil, fmt.Errorf("UPDATE without WHERE: call AllRows() to update every row")
	}

	// Il testo di SetRaw non passa da render(): è SQL scritto a mano,
	// già nella sintassi del database.
	n := 0
	var s strings.Builder
	s.WriteString(q.dialect.render("UPDATE "+q.update+" SET "+strings.Join(q.set, ", "), &n))
	if len(q.setRaw) > 0 {
		if len(q.set) > 0 {
			s.WriteString(", ")
		}
		s.WriteString(strings.Join(q.setRaw, ", "))
	}
	if q.where != "" {
		s.WriteString(q.dialect.render(" WHERE "+q.where, &n))
	}

	binds := append(append([]interface{}{}, q.setBinds...), q.whereBinds...)
	return s.String(), binds, nil
}

// Build costruisce la query finale e la esegue.
func (q *UpdateQBuilder) Build() Res {
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
	rowsAffected, _ := row.RowsAffected()
	res.Data = append(res.Data, map[string]interface{}{"rowsAffected": rowsAffected})
	return res
}

/* Example

UPDATE Customers
SET ContactName = 'Alfred Schmidt', City= 'Frankfurt'
WHERE CustomerID = 1;

	res := UpdateQuery(db, false).
		Update("clienti").
		Set("name = ?", "andrea", "email = ?", "andrea@example.com").
		Where("ID = ?", 1).
		Build()
*/
