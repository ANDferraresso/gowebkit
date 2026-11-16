package session

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MysqlStore salva le sessioni in una tabella MySQL/MariaDB.
//
// Le date di scadenza sono salvate come DATETIME nell'ora locale del server Go
// (time.Local), e rilette nello stesso fuso. Funziona sia con parseTime=false
// sia con parseTime=true nel DSN.
type MysqlStore struct {
	db        *sql.DB
	tableName string
}

// noExpiryDate è il valore salvato per una sessione senza scadenza (tm zero).
const noExpiryDate = "0001-01-01 00:00:00"

// queryRower è soddisfatto sia da *sql.DB sia da *sql.Tx.
type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

func (st *MysqlStore) CreateStore() string {
	return "CREATE TABLE `" + st.tableName + "` (\n" +
		"\t`sid` varchar(64) NOT NULL,\n" +
		"\t`data` mediumtext NOT NULL,\n" +
		"\t`expires_at` datetime NOT NULL,\n" +
		"\tPRIMARY KEY (`sid`)\n" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;"
}

// formatExpiresAt converte la scadenza nel formato DATETIME, in ora locale.
func formatExpiresAt(t time.Time) string {
	if t.IsZero() {
		return noExpiryDate
	}
	return t.In(time.Local).Format(dateFormat)
}

// parseExpiresAt interpreta il valore letto da expires_at come ora locale.
// Accetta "2006-01-02 15:04:05" (parseTime=false) e RFC3339 (parseTime=true,
// dove database/sql converte il time.Time in stringa). In entrambi i casi conta
// solo l'ora "da orologio" salvata, che viene interpretata in time.Local.
// Il secondo valore è false se la sessione non ha scadenza.
func parseExpiresAt(s string) (time.Time, bool, error) {
	if s == noExpiryDate || strings.HasPrefix(s, "0000-00-00") || strings.HasPrefix(s, "0001-01-01") {
		return time.Time{}, false, nil
	}
	if t, err := time.ParseInLocation(dateFormat, s, time.Local); err == nil {
		return t, true, nil
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, false, errors.New(ErrParsingExpiresAt)
	}
	if t.Year() <= 1 {
		return time.Time{}, false, nil
	}
	local := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	return local, true, nil
}

func isExpired(expiresAt time.Time, hasExpiry bool) bool {
	return hasExpiry && time.Now().After(expiresAt)
}

// loadRow legge dati e scadenza della sessione. Con lock=true usa
// SELECT ... FOR UPDATE (da usare dentro una transazione).
func (st *MysqlStore) loadRow(q queryRower, sid string, lock bool) (string, time.Time, bool, error) {
	var data, expiresAt string
	query := "SELECT `data`, `expires_at` FROM `" + st.tableName + "` WHERE `sid` = ?"
	if lock {
		query += " FOR UPDATE"
	}
	err := q.QueryRow(query, sid).Scan(&data, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, false, errors.New(ErrSessionNotFound)
		}
		return "", time.Time{}, false, errors.New(ErrSqlSelectQuery)
	}
	t, hasExpiry, err := parseExpiresAt(expiresAt)
	if err != nil {
		return "", time.Time{}, false, err
	}
	return data, t, hasExpiry, nil
}

func (st *MysqlStore) withTx(fn func(tx *sql.Tx) (commitAnyway bool, err error)) error {
	return withTx(st.db, fn)
}

// withTx esegue fn in una transazione: commit se fn ritorna nil o se
// commitAnyway è true (usato per salvare la cancellazione di una sessione scaduta).
// Usata da MysqlStore e PostgresStore.
func withTx(db *sql.DB, fn func(tx *sql.Tx) (commitAnyway bool, err error)) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				log.Println("[DEBUG] Error while tx.Rollback():", err)
			}
		}
	}()

	commitAnyway, fnErr := fn(tx)
	if fnErr != nil && !commitAnyway {
		return fnErr
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return fnErr
}

// modify carica la sessione con un lock sulla riga, applica fn e salva:
// richieste concorrenti sulla stessa sessione vengono serializzate.
func (st *MysqlStore) modify(sid string, fn func(SessionData) error) error {
	return st.withTx(func(tx *sql.Tx) (bool, error) {
		raw, expiresAt, hasExpiry, err := st.loadRow(tx, sid, true)
		if err != nil {
			return false, err
		}
		if isExpired(expiresAt, hasExpiry) {
			if _, err := tx.Exec("DELETE FROM `"+st.tableName+"` WHERE `sid` = ?", sid); err != nil {
				return false, errors.New(ErrSqlDeleteQuery)
			}
			return true, errors.New(ErrSessionExpired)
		}

		var data SessionData
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			return false, errors.New(ErrUnableUnserializeJson)
		}
		if data == nil {
			data = SessionData{}
		}
		if err := fn(data); err != nil {
			return false, err
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return false, errors.New(ErrUnableSerializeJson)
		}
		if _, err := tx.Exec("UPDATE `"+st.tableName+"` SET `data` = ? WHERE `sid` = ?", string(jsonData), sid); err != nil {
			return false, errors.New(ErrSqlUpdateQuery)
		}
		return false, nil
	})
}

// Inizializza lo store. db deve essere un *sql.DB.
func (st *MysqlStore) InitStore(db interface{}, tableName string) error {
	if !validTableName.MatchString(tableName) {
		return errors.New(ErrInvTableName)
	}
	sqlDB, ok := db.(*sql.DB)
	if !ok || sqlDB == nil {
		return errors.New(ErrInvStoreClient)
	}
	st.db = sqlDB
	st.tableName = tableName
	return nil
}

// Inizializza una sessione che scade a tm, restituendo il sid.
// Con tm zero la sessione non scade; con tm nel passato ritorna ErrInvExpiration.
func (st *MysqlStore) Init(tm time.Time) (string, error) {
	if !tm.IsZero() && !tm.After(time.Now()) {
		return "", errors.New(ErrInvExpiration)
	}
	sid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	_, err = st.db.Exec("INSERT INTO `"+st.tableName+"`(`sid`, `data`, `expires_at`) VALUES(?, ?, ?)",
		sid, "{}", formatExpiresAt(tm))
	if err != nil {
		return "", errors.New(ErrSqlInsertQuery)
	}
	return sid, nil
}

// Controlla se una sessione esiste e non è scaduta (dato il sid).
func (st *MysqlStore) Check(sid string) bool {
	_, expiresAt, hasExpiry, err := st.loadRow(st.db, sid, false)
	if err != nil {
		if err.Error() == ErrParsingExpiresAt {
			_ = st.Delete(sid)
		}
		return false
	}
	if isExpired(expiresAt, hasExpiry) {
		_ = st.Delete(sid)
		return false
	}
	return true
}

// Setta il valore di una chiave di una sessione (dati sid, key e value).
func (st *MysqlStore) Set(sid string, key string, value interface{}) error {
	return st.modify(sid, func(data SessionData) error {
		data[key] = value
		return nil
	})
}

// Ottiene il valore di una chiave di una sessione (dati sid e key).
func (st *MysqlStore) Get(sid string, key string) (interface{}, error) {
	raw, expiresAt, hasExpiry, err := st.loadRow(st.db, sid, false)
	if err != nil {
		return nil, err
	}
	if isExpired(expiresAt, hasExpiry) {
		_ = st.Delete(sid)
		return nil, errors.New(ErrSessionExpired)
	}
	var data SessionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, errors.New(ErrUnableUnserializeJson)
	}
	val, ok := data[key]
	if !ok {
		return nil, errors.New(ErrKeyNotFound)
	}
	return val, nil
}

// Cancella il valore collegato a una chiave di una sessione (dati sid e key).
func (st *MysqlStore) RemoveKey(sid string, key string) error {
	return st.modify(sid, func(data SessionData) error {
		if _, ok := data[key]; !ok {
			return errors.New(ErrKeyNotFound)
		}
		delete(data, key)
		return nil
	})
}

// Cancella un'intera sessione (dato il sid).
func (st *MysqlStore) Delete(sid string) error {
	_, err := st.db.Exec("DELETE FROM `"+st.tableName+"` WHERE `sid` = ?", sid)
	if err != nil {
		return errors.New(ErrSqlDeleteQuery)
	}
	return nil
}

// Sostituisce il sid della sessione con uno nuovo, mantenendo dati e scadenza
// (il vecchio sid viene eliminato). Da usare ad esempio dopo il login.
func (st *MysqlStore) Refresh(sid string) (string, error) {
	newSid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	err = st.withTx(func(tx *sql.Tx) (bool, error) {
		raw, expiresAt, hasExpiry, err := st.loadRow(tx, sid, true)
		if err != nil {
			return false, err
		}
		if _, err := tx.Exec("DELETE FROM `"+st.tableName+"` WHERE `sid` = ?", sid); err != nil {
			return false, errors.New(ErrSqlDeleteQuery)
		}
		if isExpired(expiresAt, hasExpiry) {
			return true, errors.New(ErrSessionExpired)
		}
		exp := noExpiryDate
		if hasExpiry {
			exp = formatExpiresAt(expiresAt)
		}
		if _, err := tx.Exec("INSERT INTO `"+st.tableName+"`(`sid`, `data`, `expires_at`) VALUES(?, ?, ?)", newSid, raw, exp); err != nil {
			return false, errors.New(ErrSqlInsertQuery)
		}
		return false, nil
	})
	if err != nil {
		return "", err
	}
	return newSid, nil
}
