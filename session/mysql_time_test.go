package session

import (
	"testing"
	"time"
)

// withLocal runs fn with time.Local set to a non-UTC zone, to catch
// local/UTC mix-ups in the expires_at handling.
func withLocal(t *testing.T, fn func()) {
	t.Helper()
	saved := time.Local
	time.Local = time.FixedZone("UTC+2", 2*60*60)
	defer func() { time.Local = saved }()
	fn()
}

func TestExpiresAtRoundTrip(t *testing.T) {
	withLocal(t, func() {
		exp := time.Now().Add(30 * time.Minute).Truncate(time.Second)

		// parseTime=false: the driver returns "2006-01-02 15:04:05".
		got, has, err := parseExpiresAt(formatExpiresAt(exp))
		if err != nil || !has {
			t.Fatalf("parse: %v, has=%v", err, has)
		}
		if !got.Equal(exp) {
			t.Errorf("round trip = %v, want %v (diff %v)", got, exp, got.Sub(exp))
		}
		if isExpired(got, has) {
			t.Error("session expiring in 30 minutes reported as expired")
		}

		// parseTime=true (default loc=UTC): database/sql turns the time.Time
		// into an RFC3339 string with the stored wall clock labelled as UTC.
		wall := formatExpiresAt(exp)
		asUTC, _ := time.ParseInLocation(dateFormat, wall, time.UTC)
		got, has, err = parseExpiresAt(asUTC.Format(time.RFC3339Nano))
		if err != nil || !has || !got.Equal(exp) {
			t.Errorf("parseTime=true: got %v (has=%v, err=%v), want %v", got, has, err, exp)
		}
	})
}

func TestExpiresAtPast(t *testing.T) {
	withLocal(t, func() {
		past := time.Now().Add(-time.Minute)
		got, has, err := parseExpiresAt(formatExpiresAt(past))
		if err != nil || !isExpired(got, has) {
			t.Errorf("expired session not detected: %v %v %v", got, has, err)
		}
	})
}

func TestExpiresAtNoExpiry(t *testing.T) {
	withLocal(t, func() {
		for _, s := range []string{formatExpiresAt(time.Time{}), "0000-00-00 00:00:00", "0001-01-01T00:00:00Z"} {
			_, has, err := parseExpiresAt(s)
			if err != nil || has {
				t.Errorf("parseExpiresAt(%q) = has %v, err %v; want no expiry", s, has, err)
			}
		}
		if _, _, err := parseExpiresAt("garbage"); err == nil {
			t.Error("garbage accepted")
		}
	})
}

func TestCreateStoreSQL(t *testing.T) {
	st := &MysqlStore{}
	if err := st.InitStore(nil, "sessions"); err == nil {
		t.Error("nil client accepted")
	}
	if err := st.InitStore("not a db", "sessions"); err == nil {
		t.Error("wrong client type accepted")
	}
	st.tableName = "sessions"
	want := "CREATE TABLE `sessions` (\n\t`sid` varchar(64) NOT NULL,\n\t`data` mediumtext NOT NULL,\n\t`expires_at` datetime NOT NULL,\n\tPRIMARY KEY (`sid`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
	if got := st.CreateStore(); got != want {
		t.Errorf("CreateStore() =\n%s", got)
	}
}
