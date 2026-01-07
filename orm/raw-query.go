package orm

import (
	"database/sql"
	"log"
)

// RawQuery esegue una query "raw" (tipicamente SELECT) con binds opzionali
// e ritorna i risultati come []map[colName]value (string o nil), come Build().
// binds è opzionale (0..n parametri)
func RawQuery(db *sql.DB, debug bool, query string, binds ...interface{}) Res {
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}

	rows, err := db.Query(query, binds...)
	if err != nil {
		return *ManageErr(&res, debug, err, query)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("[DEBUG] Error deferring rows.Close():", err)
		}
	}()

	cols, err := rows.Columns()
	if err != nil {
		return *ManageErr(&res, debug, err, query)
	}

	rv := make([]map[string]interface{}, 0)

	// record è un array di puntatori a RawBytes, uno per colonna
	record := make([]interface{}, len(cols))
	for i := range record {
		record[i] = new(sql.RawBytes)
	}

	for rows.Next() {
		if err := rows.Scan(record...); err != nil {
			return *ManageErr(&res, debug, err, query)
		}

		vals := map[string]interface{}{}
		for i, colName := range cols {
			raw := *record[i].(*sql.RawBytes)
			if raw == nil {
				vals[colName] = nil
			} else {
				vals[colName] = string(raw)
			}
		}
		rv = append(rv, vals)
	}

	// oOpzionale: intercetta errori di iterazione
	if err := rows.Err(); err != nil {
		return *ManageErr(&res, debug, err, query)
	}

	res.Data = rv
	return res
}
