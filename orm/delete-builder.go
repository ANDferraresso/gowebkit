package orm

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type DeleteQBuilder struct {
	db     *sql.DB
	debug  string
	delete string
	where  string
	binds  []interface{}
	err    bool
}

// Crea e ritorna un nuovo DeleteQBuilder
func DeleteQuery(db *sql.DB, debug string) *DeleteQBuilder {
	return &DeleteQBuilder{
		db:     db,
		debug:  debug,
		delete: "",
		where:  "",
		binds:  []interface{}{},
		err:    false,
	}
}

// DELETE
func (q *DeleteQBuilder) Delete(input string) *DeleteQBuilder {
	s1 := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	re := regexp.MustCompile(s1)
	if re.Match([]byte(input)) {
		q.delete = fmt.Sprintf("`%s`", input)
	} else {
		// Errore.
		q.err = true
		q.delete = ""
	}
	return q
}

// WHERE
func (q *DeleteQBuilder) Where(inputs ...interface{}) *DeleteQBuilder {
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
					cond := inputs[i].(string)
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
func (q *DeleteQBuilder) Build() Res {
	if q.err {
		return Res{Err: true, Msg: "Error while building the query.", Data: []map[string]interface{}{}}
	}
	var s strings.Builder
	if q.delete != "" {
		s.WriteString("DELETE FROM ")
		s.WriteString(q.delete)
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
		rowsAffected, err := row.RowsAffected()
		if err != nil {
			rowsAffected = int64(0)
		}
		res.Err = false
		res.Msg = ""
		res.Data = append(res.Data, map[string]interface{}{"rowsAffected": rowsAffected})
		return res
	}
}

/* Example

DELETE FROM Customers
WHERE CustomerID = 1;

	query, binds := Query().
		Delete("clienti").
		Where("ID = ?", 1).
		Build()

	fmt.Println(query)
	fmt.Println(binds)
*/
