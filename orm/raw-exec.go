package orm

import (
	"database/sql"
)

// RawExec esegue una query "raw" (tipicamente INSERT/UPDATE) con binds opzionali
// e ritorna lastInsertId e rowsAffected in Data[0].
//
// La query è passata al driver così com'è: con PostgreSQL usa i placeholder
// $1, $2, ... Inoltre i driver PostgreSQL non supportano LastInsertId()
// (vale 0): usa RawQuery con "INSERT ... RETURNING id".

func RawExec(db *sql.DB, debug bool, query string, binds ...interface{}) Res {
	res := Res{Err: false, Msg: "", Data: []map[string]interface{}{}}

	r, err := db.Exec(query, binds...)
	if err != nil {
		return *ManageErr(&res, debug, err, query)
	}

	lastInsertId, _ := r.LastInsertId()
	rowsAffected, _ := r.RowsAffected()

	res.Data = append(res.Data, map[string]interface{}{
		"lastInsertId": lastInsertId,
		"rowsAffected": rowsAffected,
	})

	return res
}
