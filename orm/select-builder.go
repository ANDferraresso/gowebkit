package orm

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type SelectQBuilder struct {
	db          *sql.DB
	debug       bool
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

var (
	selTableNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// Crea e ritorna un nuovo SelectQBuilder
func SelectQuery(db *sql.DB, debug bool) *SelectQBuilder {
	return &SelectQBuilder{
		db:          db,
		debug:       debug,
		count:       false,
		distinct:    false,
		columnNames: []string{},
		fields:      []string{},
		from:        []string{},
		where:       "",
		order:       [][2]string{},
		limit:       -1,
		offset:      -1,
		binds:       []interface{}{},
		err:         false,
		errStr:      "",
	}
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

// SELECT
func (q *SelectQBuilder) Select(fields ...string) *SelectQBuilder {
	s1b := `^[a-zA-Z_][a-zA-Z0-9_]* AS [a-zA-Z_][a-zA-Z0-9_]*$`                         // Es. ID AS id
	s2a := `^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*$`                           // Es. c.ID
	s2b := `^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]* AS [a-zA-Z_][a-zA-Z0-9_]*$` // Es. c.ID AS id
	comp_s1b := regexp.MustCompile(s1b)
	comp_s2a := regexp.MustCompile(s2a)
	comp_s2b := regexp.MustCompile(s2b)
	//
	for _, f := range fields {
		if selTableNameRegex.Match([]byte(f)) {
			q.fields = append(q.fields, fmt.Sprintf("`%s`", f))
			q.columnNames = append(q.columnNames, f)
		} else if comp_s1b.Match([]byte(f)) {
			c := strings.Split(f, " ")
			// c[0] Colonna. c[1] AS. c[2] Alias.
			q.fields = append(q.fields, fmt.Sprintf("`%s` AS `%s`", c[0], c[2]))
			q.columnNames = append(q.columnNames, c[2])
		} else if comp_s2a.Match([]byte(f)) {
			c := strings.Split(f, ".")
			// c[0] Tabella. c[1] Colonna.
			q.fields = append(q.fields, fmt.Sprintf("`%s`.`%s`", c[0], c[1]))
			q.columnNames = append(q.columnNames, c[1])
		} else if comp_s2b.Match([]byte(f)) {
			c := strings.Split(f, " ")
			// c[0] Tabella.Nome Colonna. c[1] AS. c[2] Alias.
			c_ := strings.Split(c[0], ".")
			// c_[0] Tabella. c_[1] Colonna.
			q.fields = append(q.fields, fmt.Sprintf("`%s`.`%s` AS `%s`", c_[0], c_[1], c[2]))
			q.columnNames = append(q.columnNames, c[2])
		} else {
			// Errore.
			q.err = true
			q.errStr = ""
			q.fields = []string{}
			break
		}
	}

	return q
}

// FROM
func (q *SelectQBuilder) From(inputs ...string) *SelectQBuilder {
	s2 := `^[a-zA-Z_][a-zA-Z0-9_]* AS [a-zA-Z_][a-zA-Z0-9_]*$`
	comp_s2 := regexp.MustCompile(s2)
	for _, input := range inputs {
		if selTableNameRegex.Match([]byte(input)) {
			q.from = append(q.from, fmt.Sprintf("`%s`", input))
		} else {
			if comp_s2.Match([]byte(input)) {
				c := strings.Split(input, " ")
				// c[0] Nome tabella. c[1] AS. c[2] alias.
				q.from = append(q.from, fmt.Sprintf("`%s` AS `%s`", c[0], c[2]))
			} else {
				// Errore.
				q.err = true
				q.errStr = ""
				q.from = []string{}
				break
			}
		}
	}
	return q
}

// WHERE
func (q *SelectQBuilder) Where(inputs ...interface{}) *SelectQBuilder {
	if len(inputs) > 0 {
		rxCond1a := `^[a-zA-Z_][a-zA-Z0-9_]* (<|=|>|<>|<=|>=|LIKE) \?$`
		rxCond1b := `^[a-zA-Z_][a-zA-Z0-9_]*.[a-zA-Z_][a-zA-Z0-9_]* (<|=|>|<>|<=|>=|LIKE) \?$`
		rxCond2a := `^[a-zA-Z_][a-zA-Z0-9_]* (IS NULL|IS NOT NULL)$`
		rxCond2b := `^[a-zA-Z_][a-zA-Z0-9_]*.[a-zA-Z_][a-zA-Z0-9_]* (IS NULL|IS NOT NULL)$`
		rxCond3a := `^[a-zA-Z_][a-zA-Z0-9_]* (<|=|>|<>|<=|>=) [a-zA-Z_][a-zA-Z0-9_]*$`
		rxCond3b := `^[a-zA-Z_][a-zA-Z0-9_]*.[a-zA-Z_][a-zA-Z0-9_]* (<|=|>|<>|<=|>=) [a-zA-Z_][a-zA-Z0-9_]*.[a-zA-Z_][a-zA-Z0-9_]*$`
		comp_rxCond1a := regexp.MustCompile(rxCond1a)
		comp_rxCond1b := regexp.MustCompile(rxCond1b)
		comp_rxCond2a := regexp.MustCompile(rxCond2a)
		comp_rxCond2b := regexp.MustCompile(rxCond2b)
		comp_rxCond3a := regexp.MustCompile(rxCond3a)
		comp_rxCond3b := regexp.MustCompile(rxCond3b)
		//
		rxOp := `^(AND|OR|NOT|OR NOT|AND NOT)$`
		rxBrOp := `^\(+?$`
		rxBrCl := `^\)+?$`
		i := 0
		k := 0
		for {
			if k == 0 {
				br, ok := inputs[i].(string)
				if !ok {
					q.err = true
					q.errStr = ""
					q.where = ""
					q.binds = []interface{}{}
					break
				}
				re := regexp.MustCompile(rxBrOp)
				if re.Match([]byte(br)) {
					q.where = q.where + br
					i += 1
				} else {
					cond, ok := inputs[i].(string)
					if !ok {
						q.err = true
						q.errStr = ""
						q.where = ""
						q.binds = []interface{}{}
						break
					}
					if comp_rxCond1a.Match([]byte(cond)) {
						c := strings.Split(cond, " ")
						// c[0] Nome colonna. c[1] Comparatore. c[2] ?, ma non serve.
						q.where = q.where + fmt.Sprintf("`%s` %s", c[0], c[1])
						i += 1
						k = 1
					} else if comp_rxCond1b.Match([]byte(cond)) {
						c := strings.Split(cond, " ")
						// c[0] Nome colonna. c[1] Comparatore. c[2] ?, ma non serve.
						c_ := strings.Split(c[0], ".") // c_[0] Alias. c[1] Nome colonna.
						q.where = q.where + fmt.Sprintf("`%s`.`%s` %s", c_[0], c_[1], c[1])
						i += 1
						k = 1
					} else if comp_rxCond2a.Match([]byte(cond)) {
						c := [2]string{}
						x := strings.Index(cond, " ") // Trova lo spazio
						c[0] = cond[:x]
						c[1] = cond[x+1:]
						q.where = q.where + fmt.Sprintf("`%s` %s", c[0], c[1])
						i += 1
						k = 2 // Non c'è valore dopo.
					} else if comp_rxCond2b.Match([]byte(cond)) {
						c := [2]string{}
						x := strings.Index(cond, " ") // Trova lo spazio
						c[0] = cond[:x]
						c[1] = cond[x+1:]
						c_ := strings.Split(c[0], ".") // c_[0] Alias. c[1] Nome colonna.
						q.where = q.where + fmt.Sprintf("`%s`.`%s` %s", c_[0], c_[1], c[1])
						i += 1
						k = 2 // Non c'è valore dopo.
					} else if comp_rxCond3a.Match([]byte(cond)) {
						c := strings.Split(cond, " ")
						// c[0] Nome colonna. c[1] Comparatore. c[2] Nome colonna.
						q.where = q.where + fmt.Sprintf("`%s` %s `%s`", c[0], c[1], c[2])
						i += 1
						k = 2 // Non c'è valore dopo.
					} else if comp_rxCond3b.Match([]byte(cond)) {
						c := strings.Split(cond, " ")
						// c[0] Nome colonna. c[1] Comparatore. c[2] Nome colonna.
						c0 := strings.Split(c[0], ".")
						c2 := strings.Split(c[2], ".")
						q.where = q.where + fmt.Sprintf("`%s`.`%s` %s `%s`.`%s`", c0[0], c0[1], c[1], c2[0], c2[1])
						i += 1
						k = 2 // Non c'è valore dopo.
					} else {
						// Errore.
						q.err = true
						q.errStr = ""
						q.where = ""
						q.binds = []interface{}{}
						break
					}
				}
			} else if k == 1 {
				// q.where = q.where + fmt.Sprintf(" \"%s\"", inputs[i])
				q.where = q.where + " ?"
				q.binds = append(q.binds, inputs[i])
				i += 1
				k = 2
			} else if k == 2 {
				br, ok := inputs[i].(string)
				if !ok {
					q.err = true
					q.errStr = ""
					q.where = ""
					q.binds = []interface{}{}
					break
				}
				re := regexp.MustCompile(rxBrCl)
				if re.Match([]byte(br)) {
					q.where = q.where + br
					i += 1
				} else {
					op, ok := inputs[i].(string)
					if !ok {
						q.err = true
						q.errStr = ""
						q.where = ""
						q.binds = []interface{}{}
						break
					}
					re := regexp.MustCompile(rxOp)
					if re.Match([]byte(op)) {
						q.where = q.where + fmt.Sprintf(" %s ", op)
						i += 1
						k = 0
					} else {
						// Errore.
						q.err = true
						q.errStr = ""
						q.where = ""
						q.binds = []interface{}{}
						break
					}
				}
			}

			if i >= len(inputs) {
				if k == 1 {
					// Errore.
					q.err = true
					q.errStr = ""
					q.where = ""
					q.binds = []interface{}{}
				}
				break
			}
		}
	}

	return q
}

// ORDER
func (q *SelectQBuilder) Order(order [][2]string) *SelectQBuilder {
	q.order = order
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

// Build costruisce la query finale e la esegue.
func (q *SelectQBuilder) Build() Res {
	if q.err {
		return Res{Err: true, Msg: "Error while building the query: " + q.errStr + ".", Data: []map[string]interface{}{}}
	}
	var s strings.Builder
	if q.count {
		s.WriteString("SELECT COUNT(*) AS `CNT`")
		q.columnNames = []string{"CNT"}
	} else {
		if q.distinct {
			s.WriteString("SELECT DISTINCT ")
		} else {
			s.WriteString("SELECT ")
		}
		for i, f := range q.fields {
			if i < len(q.fields)-1 {
				s.WriteString(fmt.Sprintf("%s, ", f))
			} else {
				s.WriteString(f)
			}
		}
	}
	if len(q.from) > 0 {
		s.WriteString(" FROM ")
		for i, f := range q.from {
			if i < len(q.from)-1 {
				s.WriteString(fmt.Sprintf("%s, ", f))
			} else {
				s.WriteString(f)
			}
		}
	}
	if len(q.where) > 0 {
		s.WriteString(" WHERE ")
		s.WriteString(q.where)
	}
	if len(q.order) > 0 {
		s.WriteString(" ORDER BY ")
		for i, ord := range q.order {
			if i < len(q.order)-1 {
				s.WriteString(fmt.Sprintf("`%s` %s, ", ord[0], ord[1]))
			} else {
				s.WriteString(fmt.Sprintf("`%s` %s", ord[0], ord[1]))
			}
		}
	}
	if q.limit > 0 {
		if q.offset >= 0 {
			s.WriteString(" LIMIT ")
			s.WriteString(strconv.FormatInt(q.offset, 10))
			s.WriteString(", ")
			s.WriteString(strconv.FormatInt(q.limit, 10))
		} else {
			s.WriteString(" LIMIT ")
			s.WriteString(strconv.FormatInt(q.limit, 10))
		}
	}

	// Esegui query
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}

	rv := []map[string]interface{}{}
	record := make([]interface{}, len(q.columnNames))
	for i := range record {
		record[i] = new(sql.RawBytes)
	}
	rows, err := q.db.Query(s.String(), q.binds...)
	if err != nil {
		return *ManageErr(&res, q.debug, err, s.String())
	} else {
		// defer rows.Close()
		defer func() {
			if err := rows.Close(); err != nil {
				log.Println("[DEBUG] Error deferring rows.Close():", err)
			}
		}()
		for rows.Next() {
			if err := rows.Scan(record...); err != nil {
				return *ManageErr(&res, q.debug, err, s.String())
			} else {
				vals := map[string]interface{}{}
				for i, k := range q.columnNames {
					// sql.RawBytes è un alias per []byte
					raw := *record[i].(*sql.RawBytes)
					if raw == nil {
						vals[k] = nil
					} else {
						vals[k] = string(raw)
						// vals[k] = raw
					}
				}
				rv = append(rv, vals)
			}
		}
	}

	res.Err = false
	res.Msg = ""
	res.Data = rv
	return res
}

/* Example
fields := []string{"ID", "c.name", "email"}
query, binds := Query().
	Select(fields).
	From("clienti").
	From("clienti AS boh", "fornitori").
	Where("((", "id = ?", 1, ")", "AND", "(", "c.name IS NULL", "))", "OR", "surname = ?", "Pluto", "OR", "name = ?", "Pippo").
	Order([][2]string{{"ID", "ASC"}, {"name", "ASC"}}).
	Limit(10).Offset(0).
	Build()

fmt.Println(query)
fmt.Println(binds)
*/
