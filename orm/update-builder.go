package orm

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type UpdateQBuilder struct {
	db     *sql.DB
	debug  string
	update string
	set    string
	setRaw string
	where  string
	binds  []interface{}
	err    bool
}

// Crea e ritorna un nuovo UpdateQBuilder
func UpdateQuery(db *sql.DB, debug string) *UpdateQBuilder {
	return &UpdateQBuilder{
		db:     db,
		debug:  debug,
		update: "",
		set:    "",
		setRaw: "",
		where:  "",
		binds:  []interface{}{},
		err:    false,
	}
}

// UPDATE
func (q *UpdateQBuilder) Update(input string) *UpdateQBuilder {
	s1 := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	re := regexp.MustCompile(s1)
	if re.Match([]byte(input)) {
		q.update = fmt.Sprintf("`%s`", input)
	} else {
		// Errore.
		q.err = true
		q.update = ""
	}
	return q
}

// SET
func (q *UpdateQBuilder) Set(inputs ...interface{}) *UpdateQBuilder {
	if len(inputs) > 0 {
		s1 := `^[a-zA-Z_][a-zA-Z0-9_]* = \?$`
		i := 0
		k := 0
		for {
			if k == 0 {
				input, ok := inputs[i].(string)
				if !ok {
					q.err = true
					q.where = ""
					q.binds = []interface{}{}
					break
				}
				re := regexp.MustCompile(s1)
				if re.Match([]byte(input)) {
					c := strings.Split(input, " ")
					// c[0] Nome colonna. c[1] =. c[2] ?.
					q.set = q.set + fmt.Sprintf("`%s` = ?, ", c[0])
					i += 1
					k = 1
				} else {
					// Errore.
					q.err = true
					q.set = ""
					q.binds = []interface{}{}
					break
				}
			} else if k == 1 {
				q.binds = append(q.binds, inputs[i])
				i += 1
				k = 0
			}

			if i >= len(inputs) {
				if k == 2 {
					// Errore.
					q.err = true
					q.set = ""
					q.binds = []interface{}{}
				}
				q.set = strings.TrimRight(q.set, ", ")
				break
			}
		}
	}

	return q
}

// SET_RAW
func (q *UpdateQBuilder) SetRaw(inputs ...string) *UpdateQBuilder {
	if len(inputs) > 0 {
		for _, input := range inputs {
			if len(q.setRaw) == 0 {
				q.setRaw = input
			} else {
				q.setRaw = q.setRaw + ", " + input
			}
		}
	}

	return q
}

// WHERE
func (q *UpdateQBuilder) Where(inputs ...interface{}) *UpdateQBuilder {
	if len(inputs) > 0 {
		rxCond1a := `^[a-zA-Z_][a-zA-Z0-9_]* (<|=|>|<>|<=|>=|LIKE) \?$`
		rxCond2a := `^[a-zA-Z_][a-zA-Z0-9_]* (IS NULL|IS NOT NULL)$`
		rxCond3a := `^[a-zA-Z_][a-zA-Z0-9_]* (<|=|>|<>|<=|>=) [a-zA-Z_][a-zA-Z0-9_]*$`
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
						q.where = ""
						q.binds = []interface{}{}
						break
					}
					re1a := regexp.MustCompile(rxCond1a)
					re2a := regexp.MustCompile(rxCond2a)
					re3a := regexp.MustCompile(rxCond3a)
					if re1a.Match([]byte(cond)) {
						c := strings.Split(cond, " ")
						// c[0] Nome colonna. c[1] Comparatore. c[2] ?, ma non serve.
						q.where = q.where + fmt.Sprintf("`%s` %s", c[0], c[1])
						i += 1
						k = 1
					} else if re2a.Match([]byte(cond)) {
						c := [2]string{}
						x := strings.Index(cond, " ") // Trova lo spazio
						c[0] = cond[:x]
						c[1] = cond[x+1:]
						q.where = q.where + fmt.Sprintf("`%s` %s", c[0], c[1])
						i += 1
						k = 2 // Non c'è valore dopo.
					} else if re3a.Match([]byte(cond)) {
						c := strings.Split(cond, " ")
						// c[0] Nome colonna. c[1] Comparatore. c[2] Nome colonna.
						q.where = q.where + fmt.Sprintf("`%s` %s `%s`", c[0], c[1], c[2])
						i += 1
						k = 2 // Non c'è valore dopo.
					} else {
						// Errore.
						q.err = true
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
					q.where = ""
					q.binds = []interface{}{}
				}
				break
			}
		}
	}

	return q
}

// Build costruisce la query finale e la esegue.
func (q *UpdateQBuilder) Build() Res {
	if q.err {
		return Res{Err: true, Msg: "Error while building the query.", Data: []map[string]interface{}{}}
	}
	var s strings.Builder
	if q.update != "" {
		s.WriteString("UPDATE ")
		s.WriteString(q.update)
		s.WriteString(" SET ")
		if len(q.set) > 0 {
			s.WriteString(q.set)
			if len(q.setRaw) > 0 {
				s.WriteString(", ")
				s.WriteString(q.setRaw)
			}
		} else {
			if len(q.setRaw) > 0 {
				s.WriteString(q.setRaw)
			}
		}
		s.WriteString(" ")
	}
	if len(q.where) > 0 {
		s.WriteString(" WHERE ")
		s.WriteString(q.where)
	}

	// Esegui query
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}
	row, err := q.db.Exec(s.String(), q.binds...)
	if err != nil {
		return *ManageErr(&res, q.debug, err, s.String())
	} else {
		var rowsAffected int64
		rowsAffected, _ = row.RowsAffected()
		res.Err = false
		res.Msg = ""
		res.Data = append(res.Data, map[string]interface{}{"rowsAffected": rowsAffected})
		return res
	}
}

/* Example

UPDATE Customers
SET ContactName = 'Alfred Schmidt', City= 'Frankfurt'
WHERE CustomerID = 1;

	fields := []string{"email"}
	query, binds := Query().
		Update("clienti").Set("name = ?", "andrea", "email = ?", "andrea@gmail.com").
		Where()-
		Build()

	fmt.Println(query)
	fmt.Println(binds)
*/
