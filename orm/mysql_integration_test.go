package orm

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// Integration tests against a real MySQL/MariaDB server. They run only when
// GOWEBKIT_MYSQL_DSN is set, e.g.
//
//	GOWEBKIT_MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/gowebkit_test' go test ./orm
//
// Each test creates its own table with a random name and drops it at the end.

func mysqlTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("GOWEBKIT_MYSQL_DSN")
	if dsn == "" {
		t.Skip("GOWEBKIT_MYSQL_DSN not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("cannot connect to MySQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func randomTable(t *testing.T, prefix string) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return prefix + hex.EncodeToString(b)
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...interface{}) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
}

func mustOK(t *testing.T, what string, res Res) Res {
	t.Helper()
	if res.Err {
		t.Fatalf("%s: %s", what, res.Msg)
	}
	return res
}

func TestMySQLBuilders(t *testing.T) {
	db := mysqlTestDB(t)
	table := randomTable(t, "gwk_orm_")
	mustExec(t, db, "CREATE TABLE `"+table+"` ("+
		"`id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,"+
		"`name` varchar(50) NOT NULL,"+
		"`email` varchar(100) NULL,"+
		"`active` tinyint(1) NOT NULL DEFAULT 0,"+
		"`counter` int NOT NULL DEFAULT 0"+
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	t.Cleanup(func() { _, _ = db.Exec("DROP TABLE IF EXISTS `" + table + "`") })

	// INSERT + lastInsertId
	people := []struct {
		name   string
		email  interface{}
		active bool
	}{
		{"Andrea", "andrea@example.com", true},
		{"Bruno", nil, false},
		{"Carla", "carla@example.com", true},
		{"àlice", "alice@example.com", false},
	}
	var lastID int64
	for _, p := range people {
		res := mustOK(t, "insert", InsertQuery(db, true).
			Insert(table).Columns("name", "email", "active").Values(p.name, p.email, p.active).Build())
		id, ok := res.Data[0]["lastInsertId"].(int64)
		if !ok || id <= lastID {
			t.Fatalf("lastInsertId = %#v, previous %d", res.Data[0]["lastInsertId"], lastID)
		}
		lastID = id
	}

	// RETURNING is refused by the MySQL dialect.
	if res := InsertQuery(db, true).Insert(table).Columns("name").Values("x").Returning("id").Build(); !res.Err {
		t.Error("RETURNING accepted with MySQL")
	}

	// SELECT with WHERE, groups, NULL, ORDER, LIMIT/OFFSET
	res := mustOK(t, "select", SelectQuery(db, true).
		Select("id", "t.name AS n", "email", "active").
		From(table+" AS t").
		Where("(", "t.active = ?", true, "OR", "email IS NULL", ")", "AND", "name <> ?", "Carla").
		Order([][2]string{{"name", "ASC"}}).
		Build())
	if len(res.Data) != 2 || res.Data[0]["n"] != "Andrea" || res.Data[1]["n"] != "Bruno" {
		t.Fatalf("select = %v", res.Data)
	}
	if res.Data[1]["email"] != nil {
		t.Errorf("NULL email = %#v, want nil", res.Data[1]["email"])
	}
	if res.Data[0]["active"] != "1" {
		t.Errorf("active = %#v, want \"1\"", res.Data[0]["active"])
	}

	res = mustOK(t, "limit/offset", SelectQuery(db, true).Select("name").From(table).
		Order([][2]string{{"id", ""}}).Limit(2).Offset(1).Build())
	if len(res.Data) != 2 || res.Data[0]["name"] != "Bruno" || res.Data[1]["name"] != "Carla" {
		t.Errorf("limit 2 offset 1 = %v", res.Data)
	}

	res = mustOK(t, "count", SelectQuery(db, true).Count().From(table).Where("active = ?", 1).Build())
	if res.Data[0]["CNT"] != "2" {
		t.Errorf("count = %v", res.Data)
	}

	// ILIKE -> LIKE, case-insensitive with the default _ci collation.
	res = mustOK(t, "ilike", SelectQuery(db, true).Select("name").From(table).Where("name ILIKE ?", "AND%").Build())
	if len(res.Data) != 1 || res.Data[0]["name"] != "Andrea" {
		t.Errorf("ILIKE = %v", res.Data)
	}

	// UPDATE: Where() before Set(), SET binds first, SetRaw.
	res = mustOK(t, "update", UpdateQuery(db, true).
		Update(table).
		Where("name = ?", "Bruno").
		Set("email = ?", "bruno@example.com").
		SetRaw("`counter` = `counter` + 5").
		Build())
	if res.Data[0]["rowsAffected"] != int64(1) {
		t.Errorf("update rowsAffected = %v", res.Data)
	}
	res = mustOK(t, "check update", RawQuery(db, true, "SELECT `email`, `counter` FROM `"+table+"` WHERE `name` = ?", "Bruno"))
	if len(res.Data) != 1 || res.Data[0]["email"] != "bruno@example.com" || res.Data[0]["counter"] != "5" {
		t.Errorf("after update = %v", res.Data)
	}

	if res := UpdateQuery(db, true).Update(table).Set("counter = ?", 0).Build(); !res.Err {
		t.Error("UPDATE without WHERE executed")
	}
	res = mustOK(t, "update all", UpdateQuery(db, true).Update(table).Set("counter = ?", 1).AllRows().Build())
	if res.Data[0]["rowsAffected"] != int64(len(people)) {
		t.Errorf("update all rowsAffected = %v", res.Data)
	}

	// RawExec
	res = mustOK(t, "raw exec", RawExec(db, true, "UPDATE `"+table+"` SET `counter` = `counter` + 1 WHERE `id` = ?", lastID))
	if res.Data[0]["rowsAffected"] != int64(1) {
		t.Errorf("raw exec = %v", res.Data)
	}

	// DELETE
	if res := DeleteQuery(db, true).Delete(table).Build(); !res.Err {
		t.Error("DELETE without WHERE executed")
	}
	res = mustOK(t, "delete", DeleteQuery(db, true).Delete(table).Where("email IS NULL", "OR", "name = ?", "Carla").Build())
	// No NULL emails are left (Bruno's was updated): only Carla is deleted.
	if res.Data[0]["rowsAffected"] != int64(1) {
		t.Errorf("delete rowsAffected = %v", res.Data)
	}
	mustOK(t, "delete all", DeleteQuery(db, true).Delete(table).AllRows().Build())
	res = mustOK(t, "count after delete", SelectQuery(db, true).Count().From(table).Build())
	if res.Data[0]["CNT"] != "0" {
		t.Errorf("rows left = %v", res.Data)
	}

	// DBMS errors are reported, not panics.
	if res := SelectQuery(db, false).Select("missing").From(table).Build(); !res.Err || res.Msg != "DBMS ERROR." {
		t.Errorf("bad column: %+v", res)
	}
}
