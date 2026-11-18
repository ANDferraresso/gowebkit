package orm

import "testing"

// Conditions that look qualified but don't use a dot must be rejected,
// not crash with index out of range.
func TestSelectWhereRejectsNonDotQualifier(t *testing.T) {
	conds := []string{
		"a-b = ?",
		"a`b = ?",
		"a-b IS NULL",
		"a-b = c-d",
	}
	for _, c := range conds {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Where(%q) panicked: %v", c, r)
				}
			}()
			args := []interface{}{c}
			if c == "a-b = ?" || c == "a`b = ?" {
				args = append(args, 1)
			}
			q := SelectQuery(nil, false).Where(args...)
			if !q.err {
				t.Errorf("Where(%q) accepted, want error", c)
			}
		}()
	}
}

func TestSelectWhereQualified(t *testing.T) {
	q := SelectQuery(nil, false).Where("c.id = ?", 1, "AND", "c.name IS NULL", "OR", "c.a = d.b")
	if q.err {
		t.Fatal("unexpected error")
	}
	want := "`c`.`id` = ? AND `c`.`name` IS NULL OR `c`.`a` = `d`.`b`"
	if q.where != want {
		t.Errorf("where = %q, want %q", q.where, want)
	}
}

func TestSelectOrderDirection(t *testing.T) {
	valid := []struct {
		in, want string
	}{
		{"ASC", "ASC"},
		{"desc", "DESC"},
		{" Desc ", "DESC"},
		{"", "ASC"},
	}
	for _, tt := range valid {
		q := SelectQuery(nil, false).Order([][2]string{{"name", tt.in}})
		if q.err {
			t.Errorf("Order direction %q rejected", tt.in)
			continue
		}
		if q.order[0][1] != tt.want {
			t.Errorf("Order direction %q = %q, want %q", tt.in, q.order[0][1], tt.want)
		}
	}

	invalid := []string{"ASC; DROP TABLE users", "ASC, (SELECT 1)", "UP"}
	for _, d := range invalid {
		q := SelectQuery(nil, false).Order([][2]string{{"name", d}})
		if !q.err {
			t.Errorf("Order direction %q accepted, want error", d)
		}
	}
}

func TestSelectOrderDoesNotMutateInput(t *testing.T) {
	order := [][2]string{{"c.name", "asc"}}
	q := SelectQuery(nil, false).Order(order)
	if q.err {
		t.Fatal("unexpected error")
	}
	if order[0][0] != "c.name" || order[0][1] != "asc" {
		t.Errorf("caller slice modified: %v", order)
	}
	if q.order[0][0] != "`c`.`name`" {
		t.Errorf("order column = %q", q.order[0][0])
	}
}
