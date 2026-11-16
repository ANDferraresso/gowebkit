package session

import (
	"database/sql"
	"testing"
	"time"
)

// Senza un server PostgreSQL si verificano solo la parte che non tocca il DB.
func TestPostgresStoreNoDB(t *testing.T) {
	st := &PostgresStore{}
	if err := st.InitStore(&sql.DB{}, "bad-name"); err == nil || err.Error() != ErrInvTableName {
		t.Errorf("invalid table name: err = %v", err)
	}
	if err := st.InitStore("not a db", "sessions"); err == nil || err.Error() != ErrInvStoreClient {
		t.Errorf("invalid client: err = %v", err)
	}
	if err := st.InitStore((*sql.DB)(nil), "sessions"); err == nil || err.Error() != ErrInvStoreClient {
		t.Errorf("nil client: err = %v", err)
	}
	if err := st.InitStore(&sql.DB{}, "sessions"); err != nil {
		t.Fatal(err)
	}

	want := "CREATE TABLE \"sessions\" (\n" +
		"\t\"sid\" varchar(64) NOT NULL PRIMARY KEY,\n" +
		"\t\"data\" text NOT NULL,\n" +
		"\t\"expires_at\" timestamptz NULL\n" +
		");"
	if got := st.CreateStore(); got != want {
		t.Errorf("CreateStore =\n%s\nwant\n%s", got, want)
	}

	if _, err := st.Init(time.Now().Add(-time.Minute)); err == nil || err.Error() != ErrInvExpiration {
		t.Errorf("past expiration: err = %v", err)
	}

	if v := pgExpiresAt(time.Time{}); v.Valid {
		t.Error("zero time must be stored as NULL")
	}
	if pgIsExpired(sql.NullTime{}) {
		t.Error("NULL expires_at must never expire")
	}
	if !pgIsExpired(sql.NullTime{Time: time.Now().Add(-time.Second), Valid: true}) {
		t.Error("past expires_at must be expired")
	}
	if pgIsExpired(sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true}) {
		t.Error("future expires_at must not be expired")
	}
}
