package orm

import (
	"reflect"
	"testing"
)

func TestParseWhere(t *testing.T) {
	tests := []struct {
		name      string
		in        []interface{}
		qualified bool
		want      string
		binds     []interface{}
		wantErr   bool
	}{
		{"simple", []interface{}{"id = ?", 1}, false, "`id` = ?", []interface{}{1}, false},
		{"groups", []interface{}{"((", "id = ?", 1, ")", "AND", "(", "name IS NULL", "))", "OR", "surname = ?", "Pluto"}, false,
			"((`id` = ?) AND (`name` IS NULL)) OR `surname` = ?", []interface{}{1, "Pluto"}, false},
		{"column compare", []interface{}{"a >= b"}, false, "`a` >= `b`", []interface{}{}, false},
		{"qualified", []interface{}{"c.id LIKE ?", "x%", "AND NOT", "c.a <> d.b"}, true,
			"`c`.`id` LIKE ? AND NOT `c`.`a` <> `d`.`b`", []interface{}{"x%"}, false},
		{"qualified not allowed", []interface{}{"c.id = ?", 1}, false, "", nil, true},
		{"empty", []interface{}{}, false, "", []interface{}{}, false},
		{"missing value", []interface{}{"id = ?"}, false, "", nil, true},
		{"trailing operator", []interface{}{"id = ?", 1, "AND"}, false, "", nil, true},
		{"unbalanced open", []interface{}{"(", "id = ?", 1}, false, "", nil, true},
		{"unbalanced close", []interface{}{"id = ?", 1, ")"}, false, "", nil, true},
		{"bad operator", []interface{}{"id = ?", 1, "XOR", "a = ?", 2}, false, "", nil, true},
		{"injection in condition", []interface{}{"id = 1 OR 1=1"}, false, "", nil, true},
		{"non-string condition", []interface{}{42}, false, "", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, binds, err := parseWhere(tt.in, tt.qualified)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("where = %q, want %q", got, tt.want)
			}
			if !reflect.DeepEqual(binds, tt.binds) {
				t.Errorf("binds = %v, want %v", binds, tt.binds)
			}
		})
	}
}

func TestSelectToSQL(t *testing.T) {
	q := SelectQuery(nil, false).
		Distinct().
		Select("ID", "name AS n", "c.email", "c.city AS town").
		From("clienti AS c", "ordini").
		Where("c.ID = ?", 5, "OR", "name IS NOT NULL").
		Order([][2]string{{"ID", "desc"}, {"c.name", ""}}).
		Limit(10).Offset(20)

	query, binds, cols, err := q.toSQL()
	if err != nil {
		t.Fatal(err)
	}
	want := "SELECT DISTINCT `ID`, `name` AS `n`, `c`.`email`, `c`.`city` AS `town` FROM `clienti` AS `c`, `ordini` WHERE `c`.`ID` = ? OR `name` IS NOT NULL ORDER BY `ID` DESC, `c`.`name` ASC LIMIT 10 OFFSET 20"
	if query != want {
		t.Errorf("query =\n%s\nwant\n%s", query, want)
	}
	if !reflect.DeepEqual(binds, []interface{}{5}) {
		t.Errorf("binds = %v", binds)
	}
	if !reflect.DeepEqual(cols, []string{"ID", "n", "email", "town"}) {
		t.Errorf("columns = %v", cols)
	}

	query, _, cols, err = SelectQuery(nil, false).Count().From("clienti").toSQL()
	if err != nil || query != "SELECT COUNT(*) AS `CNT` FROM `clienti`" || !reflect.DeepEqual(cols, []string{"CNT"}) {
		t.Errorf("count query = %q, cols %v, err %v", query, cols, err)
	}

	limits := []struct {
		limit, offset int64
		want          string
	}{
		{10, -1, "SELECT `a` FROM `t` LIMIT 10"},
		{10, 0, "SELECT `a` FROM `t` LIMIT 10"},
		{10, 5, "SELECT `a` FROM `t` LIMIT 10 OFFSET 5"},
		{0, 5, "SELECT `a` FROM `t`"},  // no limit: offset ignored, as before
		{-1, 5, "SELECT `a` FROM `t`"}, // idem
	}
	for _, tt := range limits {
		query, _, _, err := SelectQuery(nil, false).Select("a").From("t").Limit(tt.limit).Offset(tt.offset).toSQL()
		if err != nil || query != tt.want {
			t.Errorf("Limit(%d).Offset(%d) = %q, err %v; want %q", tt.limit, tt.offset, query, err, tt.want)
		}
	}

	if _, _, _, err := SelectQuery(nil, false).Select("a; DROP").From("t").toSQL(); err == nil {
		t.Error("invalid field accepted")
	}
	if _, _, _, err := SelectQuery(nil, false).Select("a").toSQL(); err == nil {
		t.Error("missing FROM accepted")
	}
}

func TestInsertToSQL(t *testing.T) {
	query, binds, err := InsertQuery(nil, false).Insert("clienti").Columns("name", "email").Values("Andrea", nil).toSQL()
	if err != nil {
		t.Fatal(err)
	}
	if query != "INSERT INTO `clienti`(`name`, `email`) VALUES(?, ?)" {
		t.Errorf("query = %q", query)
	}
	if !reflect.DeepEqual(binds, []interface{}{"Andrea", nil}) {
		t.Errorf("binds = %v", binds)
	}
	if _, _, err := InsertQuery(nil, false).Insert("t").Columns("a", "b").Values(1).toSQL(); err == nil {
		t.Error("column/value count mismatch accepted")
	}
	if _, _, err := InsertQuery(nil, false).Insert("t`x").Columns("a").Values(1).toSQL(); err == nil {
		t.Error("invalid table accepted")
	}
}

func TestUpdateToSQL(t *testing.T) {
	// Where() before Set(): SET binds must still come first.
	query, binds, err := UpdateQuery(nil, false).
		Update("clienti").
		Where("ID = ?", 7).
		Set("name = ?", "Andrea", "email = ?", "a@example.com").
		SetRaw("`counter` = `counter` + 1").
		toSQL()
	if err != nil {
		t.Fatal(err)
	}
	want := "UPDATE `clienti` SET `name` = ?, `email` = ?, `counter` = `counter` + 1 WHERE `ID` = ?"
	if query != want {
		t.Errorf("query =\n%s\nwant\n%s", query, want)
	}
	if !reflect.DeepEqual(binds, []interface{}{"Andrea", "a@example.com", 7}) {
		t.Errorf("binds = %v", binds)
	}

	if _, _, err := UpdateQuery(nil, false).Update("t").Set("a = ?").Where("id = ?", 1).toSQL(); err == nil {
		t.Error("odd SET arguments accepted")
	}
	if _, _, err := UpdateQuery(nil, false).Update("t").Set("a = ?", 1).toSQL(); err == nil {
		t.Error("UPDATE without WHERE accepted")
	}
	query, _, err = UpdateQuery(nil, false).Update("t").Set("a = ?", 1).AllRows().toSQL()
	if err != nil || query != "UPDATE `t` SET `a` = ?" {
		t.Errorf("AllRows update = %q, err %v", query, err)
	}
	if _, _, err := UpdateQuery(nil, false).Update("t").Where("id = ?", 1).toSQL(); err == nil {
		t.Error("UPDATE without SET accepted")
	}
}

func TestDeleteToSQL(t *testing.T) {
	query, binds, err := DeleteQuery(nil, false).Delete("clienti").Where("ID = ?", 1).toSQL()
	if err != nil || query != "DELETE FROM `clienti` WHERE `ID` = ?" || !reflect.DeepEqual(binds, []interface{}{1}) {
		t.Errorf("delete = %q %v, err %v", query, binds, err)
	}
	if _, _, err := DeleteQuery(nil, false).Delete("clienti").Where().toSQL(); err == nil {
		t.Error("DELETE without WHERE accepted")
	}
	query, _, err = DeleteQuery(nil, false).Delete("clienti").AllRows().toSQL()
	if err != nil || query != "DELETE FROM `clienti`" {
		t.Errorf("AllRows delete = %q, err %v", query, err)
	}
}

func TestPostgresDialect(t *testing.T) {
	query, binds, _, err := SelectQuery(nil, false).Dialect(Postgres).
		Select("id", "c.name AS n").
		From("clienti AS c").
		Where("c.id > ?", 5, "AND", "(", "name LIKE ?", "a%", "OR", "c.email IS NULL", ")").
		Order([][2]string{{"id", "desc"}}).
		Limit(10).Offset(20).
		toSQL()
	want := `SELECT "id", "c"."name" AS "n" FROM "clienti" AS "c" WHERE "c"."id" > $1 AND ("name" LIKE $2 OR "c"."email" IS NULL) ORDER BY "id" DESC LIMIT 10 OFFSET 20`
	if err != nil || query != want {
		t.Errorf("select =\n%s\nwant\n%s\nerr %v", query, want, err)
	}
	if !reflect.DeepEqual(binds, []interface{}{5, "a%"}) {
		t.Errorf("binds = %v", binds)
	}

	query, _, err = InsertQuery(nil, false).Dialect(Postgres).
		Insert("clienti").Columns("name", "email").Values("Andrea", nil).Returning("id").toSQL()
	if want := `INSERT INTO "clienti"("name", "email") VALUES($1, $2) RETURNING "id"`; err != nil || query != want {
		t.Errorf("insert = %q, err %v; want %q", query, err, want)
	}
	if _, _, err := InsertQuery(nil, false).Insert("t").Columns("a").Values(1).Returning("id").toSQL(); err == nil {
		t.Error("RETURNING accepted with the MySQL dialect")
	}

	// Numbering continues from SET to WHERE; SetRaw is left untouched.
	query, _, err = UpdateQuery(nil, false).Dialect(Postgres).
		Update("clienti").
		Where("id = ?", 7, "OR", "id = ?", 8).
		Set("name = ?", "Andrea").
		SetRaw(`"counter" = "counter" + 1`).
		toSQL()
	if want := `UPDATE "clienti" SET "name" = $1, "counter" = "counter" + 1 WHERE "id" = $2 OR "id" = $3`; err != nil || query != want {
		t.Errorf("update = %q, err %v; want %q", query, err, want)
	}
	query, _, err = UpdateQuery(nil, false).Dialect(Postgres).Update("t").SetRaw(`"n" = 0`).Where("id = ?", 1).toSQL()
	if want := `UPDATE "t" SET "n" = 0 WHERE "id" = $1`; err != nil || query != want {
		t.Errorf("update raw only = %q, err %v; want %q", query, err, want)
	}

	query, _, err = DeleteQuery(nil, false).Dialect(Postgres).Delete("clienti").Where("id = ?", 1).toSQL()
	if want := `DELETE FROM "clienti" WHERE "id" = $1`; err != nil || query != want {
		t.Errorf("delete = %q, err %v; want %q", query, err, want)
	}

	if _, _, err := DeleteQuery(nil, false).Dialect(Dialect(99)).Delete("t").AllRows().toSQL(); err == nil {
		t.Error("invalid dialect accepted")
	}
}

func TestDefaultDialect(t *testing.T) {
	old := DefaultDialect
	DefaultDialect = Postgres
	defer func() { DefaultDialect = old }()

	query, _, err := DeleteQuery(nil, false).Delete("t").Where("id = ?", 1).toSQL()
	if err != nil || query != `DELETE FROM "t" WHERE "id" = $1` {
		t.Errorf("delete = %q, err %v", query, err)
	}
	query, _, err = DeleteQuery(nil, false).Dialect(MySQL).Delete("t").Where("id = ?", 1).toSQL()
	if err != nil || query != "DELETE FROM `t` WHERE `id` = ?" {
		t.Errorf("explicit MySQL delete = %q, err %v", query, err)
	}
}

func TestILike(t *testing.T) {
	q := SelectQuery(nil, false).Select("name").From("t").Where("name ILIKE ?", "an%", "AND", "c.city ILIKE ?", "ro%")
	query, _, _, err := q.toSQL()
	if want := "SELECT `name` FROM `t` WHERE `name` LIKE ? AND `c`.`city` LIKE ?"; err != nil || query != want {
		t.Errorf("mysql = %q, err %v; want %q", query, err, want)
	}
	query, _, _, err = q.Dialect(Postgres).toSQL()
	if want := `SELECT "name" FROM "t" WHERE "name" ILIKE $1 AND "c"."city" ILIKE $2`; err != nil || query != want {
		t.Errorf("postgres = %q, err %v; want %q", query, err, want)
	}
	// A column named ILIKE is quoted and must not be rewritten.
	query, _, err = DeleteQuery(nil, false).Delete("t").Where("ILIKE = ?", 1).toSQL()
	if want := "DELETE FROM `t` WHERE `ILIKE` = ?"; err != nil || query != want {
		t.Errorf("column ILIKE = %q, err %v; want %q", query, err, want)
	}
}
