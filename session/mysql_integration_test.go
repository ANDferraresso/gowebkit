package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Integration tests for MysqlStore against a real MySQL/MariaDB server.
// They run only when GOWEBKIT_MYSQL_DSN is set, e.g.
//
//	GOWEBKIT_MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/gowebkit_test' go test ./session
//
// Every scenario runs both with parseTime=false and parseTime=true, on its own
// table with a random name, dropped at the end.

func mysqlStores(t *testing.T) map[string]*MysqlStore {
	t.Helper()
	dsn := os.Getenv("GOWEBKIT_MYSQL_DSN")
	if dsn == "" {
		t.Skip("GOWEBKIT_MYSQL_DSN not set")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("invalid GOWEBKIT_MYSQL_DSN: %v", err)
	}

	stores := map[string]*MysqlStore{}
	for _, parseTime := range []bool{false, true} {
		c := cfg.Clone()
		c.ParseTime = parseTime
		db, err := sql.Open("mysql", c.FormatDSN())
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Ping(); err != nil {
			t.Fatalf("cannot connect to MySQL: %v", err)
		}

		b := make([]byte, 6)
		_, _ = rand.Read(b)
		st := &MysqlStore{}
		if err := st.InitStore(db, "gwk_sess_"+hex.EncodeToString(b)); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(st.CreateStore()); err != nil {
			t.Fatalf("CreateStore: %v\n%s", err, st.CreateStore())
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DROP TABLE IF EXISTS `" + st.tableName + "`")
			_ = db.Close()
		})
		stores[fmt.Sprintf("parseTime=%v", parseTime)] = st
	}
	return stores
}

// setExpiresAt forces the expiration of a session, bypassing Init's checks.
func setExpiresAt(t *testing.T, st *MysqlStore, sid string, tm time.Time) {
	t.Helper()
	if _, err := st.db.Exec("UPDATE `"+st.tableName+"` SET `expires_at` = ? WHERE `sid` = ?", formatExpiresAt(tm), sid); err != nil {
		t.Fatal(err)
	}
}

func TestMysqlStoreIntegration(t *testing.T) {
	for name, st := range mysqlStores(t) {
		t.Run(name, func(t *testing.T) {
			exp := time.Now().Add(2 * time.Hour).Truncate(time.Second)
			sid, err := st.Init(exp)
			if err != nil {
				t.Fatal(err)
			}
			if !st.Check(sid) {
				t.Fatal("new session not found")
			}

			// The expiration round-trips exactly (same instant, local time).
			_, got, hasExpiry, err := st.loadRow(st.db, sid, false)
			if err != nil || !hasExpiry || !got.Equal(exp) {
				t.Errorf("expires_at = %v (hasExpiry %v, err %v), want %v", got, hasExpiry, err, exp)
			}

			// Set / Get / RemoveKey
			if err := st.Set(sid, "userID", 42); err != nil {
				t.Fatal(err)
			}
			if err := st.Set(sid, "name", "Andrea"); err != nil {
				t.Fatal(err)
			}
			if v, err := st.Get(sid, "userID"); err != nil || v != float64(42) {
				t.Errorf("Get userID = %#v, %v", v, err)
			}
			if err := st.RemoveKey(sid, "userID"); err != nil {
				t.Fatal(err)
			}
			if _, err := st.Get(sid, "userID"); err == nil || err.Error() != ErrKeyNotFound {
				t.Errorf("removed key: err = %v", err)
			}
			if err := st.RemoveKey(sid, "userID"); err == nil || err.Error() != ErrKeyNotFound {
				t.Errorf("remove missing key: err = %v", err)
			}
			if err := st.Set("no-such-sid", "k", 1); err == nil || err.Error() != ErrSessionNotFound {
				t.Errorf("Set on missing session: err = %v", err)
			}

			// Refresh keeps data and expiration, invalidates the old sid.
			newSid, err := st.Refresh(sid)
			if err != nil {
				t.Fatal(err)
			}
			if st.Check(sid) {
				t.Error("old sid still valid after Refresh")
			}
			if v, err := st.Get(newSid, "name"); err != nil || v != "Andrea" {
				t.Errorf("Get after Refresh = %#v, %v", v, err)
			}
			if _, got, _, _ := st.loadRow(st.db, newSid, false); !got.Equal(exp) {
				t.Errorf("expiration after Refresh = %v, want %v", got, exp)
			}

			// Session without expiry.
			forever, err := st.Init(time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			if !st.Check(forever) {
				t.Error("session without expiry not valid")
			}
			fresh, err := st.Refresh(forever)
			if err != nil || !st.Check(fresh) {
				t.Errorf("Refresh of session without expiry: %v", err)
			}

			// Expired session: Get reports it and deletes it.
			expired, err := st.Init(exp)
			if err != nil {
				t.Fatal(err)
			}
			setExpiresAt(t, st, expired, time.Now().Add(-time.Minute))
			if _, err := st.Get(expired, "name"); err == nil || err.Error() != ErrSessionExpired {
				t.Errorf("Get expired: err = %v", err)
			}
			if _, _, _, err := st.loadRow(st.db, expired, false); err == nil || err.Error() != ErrSessionNotFound {
				t.Errorf("expired session not deleted: err = %v", err)
			}

			// Set and Refresh on an expired session.
			expired, _ = st.Init(exp)
			setExpiresAt(t, st, expired, time.Now().Add(-time.Minute))
			if err := st.Set(expired, "k", 1); err == nil || err.Error() != ErrSessionExpired {
				t.Errorf("Set expired: err = %v", err)
			}
			expired, _ = st.Init(exp)
			setExpiresAt(t, st, expired, time.Now().Add(-time.Minute))
			if _, err := st.Refresh(expired); err == nil || err.Error() != ErrSessionExpired {
				t.Errorf("Refresh expired: err = %v", err)
			}

			// DeleteExpired removes only expired sessions.
			var old []string
			for i := 0; i < 3; i++ {
				s, _ := st.Init(exp)
				setExpiresAt(t, st, s, time.Now().Add(-time.Hour))
				old = append(old, s)
			}
			n, err := st.DeleteExpired()
			if err != nil || n != int64(len(old)) {
				t.Errorf("DeleteExpired = %d, %v; want %d", n, err, len(old))
			}
			if !st.Check(newSid) || !st.Check(fresh) {
				t.Error("DeleteExpired removed valid sessions")
			}

			// Delete
			if err := st.Delete(newSid); err != nil || st.Check(newSid) {
				t.Errorf("Delete: %v", err)
			}
		})
	}
}

// Concurrent Set on the same session must not lose updates (row lock).
func TestMysqlStoreConcurrentSet(t *testing.T) {
	for name, st := range mysqlStores(t) {
		t.Run(name, func(t *testing.T) {
			sid, err := st.Init(time.Now().Add(time.Hour))
			if err != nil {
				t.Fatal(err)
			}
			const n = 20
			var wg sync.WaitGroup
			errs := make(chan error, n)
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					errs <- st.Set(sid, fmt.Sprintf("k%d", i), i)
				}(i)
			}
			wg.Wait()
			close(errs)
			for err := range errs {
				if err != nil {
					t.Errorf("Set: %v", err)
				}
			}
			for i := 0; i < n; i++ {
				if v, err := st.Get(sid, fmt.Sprintf("k%d", i)); err != nil || v != float64(i) {
					t.Errorf("k%d = %#v, %v (update lost)", i, v, err)
				}
			}
		})
	}
}
