package session

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// PostgresStore salva le sessioni in una tabella PostgreSQL.
//
// Il pacchetto non importa alcun driver: registra tu quello che preferisci,
// ad esempio
//
//	import _ "github.com/jackc/pgx/v5/stdlib" // sql.Open("pgx", dsn)
//	import _ "github.com/lib/pq"              // sql.Open("postgres", dsn)
//
// La scadenza è salvata in una colonna timestamptz (un istante assoluto, senza
// ambiguità di fuso orario); NULL significa "nessuna scadenza".
type PostgresStore struct {
	db        *sql.DB
	tableName string
}

var (
	_ Store          = (*PostgresStore)(nil)
	_ ExpiredCleaner = (*PostgresStore)(nil)
)

func (st *PostgresStore) CreateStore() string {
	return `CREATE TABLE "` + st.tableName + `" (` + "\n" +
		"\t\"sid\" varchar(64) NOT NULL PRIMARY KEY,\n" +
		"\t\"data\" text NOT NULL,\n" +
		"\t\"expires_at\" timestamptz NULL\n" +
		");"
}

func (st *PostgresStore) table() string {
	return `"` + st.tableName + `"`
}

// pgExpiresAt converte la scadenza nel valore da salvare (NULL se tm è zero).
func pgExpiresAt(tm time.Time) sql.NullTime {
	return sql.NullTime{Time: tm, Valid: !tm.IsZero()}
}

// loadRow legge dati e scadenza della sessione. Con lock=true usa
// SELECT ... FOR UPDATE (da usare dentro una transazione).
func (st *PostgresStore) loadRow(q queryRower, sid string, lock bool) (string, sql.NullTime, error) {
	var data string
	var expiresAt sql.NullTime
	query := `SELECT "data", "expires_at" FROM ` + st.table() + ` WHERE "sid" = $1`
	if lock {
		query += " FOR UPDATE"
	}
	if err := q.QueryRow(query, sid).Scan(&data, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.NullTime{}, errors.New(ErrSessionNotFound)
		}
		return "", sql.NullTime{}, errors.New(ErrSqlSelectQuery)
	}
	return data, expiresAt, nil
}

func pgIsExpired(expiresAt sql.NullTime) bool {
	return isExpired(expiresAt.Time, expiresAt.Valid)
}

// modify carica la sessione con un lock sulla riga, applica fn e salva:
// richieste concorrenti sulla stessa sessione vengono serializzate.
func (st *PostgresStore) modify(sid string, fn func(SessionData) error) error {
	return withTx(st.db, func(tx *sql.Tx) (bool, error) {
		raw, expiresAt, err := st.loadRow(tx, sid, true)
		if err != nil {
			return false, err
		}
		if pgIsExpired(expiresAt) {
			if _, err := tx.Exec(`DELETE FROM `+st.table()+` WHERE "sid" = $1`, sid); err != nil {
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
		if _, err := tx.Exec(`UPDATE `+st.table()+` SET "data" = $1 WHERE "sid" = $2`, string(jsonData), sid); err != nil {
			return false, errors.New(ErrSqlUpdateQuery)
		}
		return false, nil
	})
}

// Inizializza lo store. db deve essere un *sql.DB aperto con un driver PostgreSQL.
func (st *PostgresStore) InitStore(db interface{}, tableName string) error {
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
func (st *PostgresStore) Init(tm time.Time) (string, error) {
	if !tm.IsZero() && !tm.After(time.Now()) {
		return "", errors.New(ErrInvExpiration)
	}
	sid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	_, err = st.db.Exec(`INSERT INTO `+st.table()+`("sid", "data", "expires_at") VALUES($1, $2, $3)`,
		sid, "{}", pgExpiresAt(tm))
	if err != nil {
		return "", errors.New(ErrSqlInsertQuery)
	}
	return sid, nil
}

// Controlla se una sessione esiste e non è scaduta (dato il sid).
func (st *PostgresStore) Check(sid string) bool {
	_, expiresAt, err := st.loadRow(st.db, sid, false)
	if err != nil {
		return false
	}
	if pgIsExpired(expiresAt) {
		_ = st.Delete(sid)
		return false
	}
	return true
}

// Setta il valore di una chiave di una sessione (dati sid, key e value).
func (st *PostgresStore) Set(sid string, key string, value interface{}) error {
	return st.modify(sid, func(data SessionData) error {
		data[key] = value
		return nil
	})
}

// Ottiene il valore di una chiave di una sessione (dati sid e key).
func (st *PostgresStore) Get(sid string, key string) (interface{}, error) {
	raw, expiresAt, err := st.loadRow(st.db, sid, false)
	if err != nil {
		return nil, err
	}
	if pgIsExpired(expiresAt) {
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
func (st *PostgresStore) RemoveKey(sid string, key string) error {
	return st.modify(sid, func(data SessionData) error {
		if _, ok := data[key]; !ok {
			return errors.New(ErrKeyNotFound)
		}
		delete(data, key)
		return nil
	})
}

// Cancella un'intera sessione (dato il sid).
func (st *PostgresStore) Delete(sid string) error {
	if _, err := st.db.Exec(`DELETE FROM `+st.table()+` WHERE "sid" = $1`, sid); err != nil {
		return errors.New(ErrSqlDeleteQuery)
	}
	return nil
}

func (st *PostgresStore) deleteExpiredQuery(now time.Time) (string, []interface{}) {
	// NULL < $1 non è vero: le sessioni senza scadenza non vengono toccate.
	return `DELETE FROM ` + st.table() + ` WHERE "expires_at" < $1`, []interface{}{now}
}

// DeleteExpired cancella tutte le sessioni scadute (vedi ExpiredCleaner).
func (st *PostgresStore) DeleteExpired() (int64, error) {
	query, args := st.deleteExpiredQuery(time.Now())
	r, err := st.db.Exec(query, args...)
	if err != nil {
		return 0, errors.New(ErrSqlDeleteQuery)
	}
	n, _ := r.RowsAffected()
	return n, nil
}

// Sostituisce il sid della sessione con uno nuovo, mantenendo dati e scadenza.
// Da usare ad esempio dopo il login.
func (st *PostgresStore) Refresh(sid string) (string, error) {
	newSid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	err = withTx(st.db, func(tx *sql.Tx) (bool, error) {
		_, expiresAt, err := st.loadRow(tx, sid, true)
		if err != nil {
			return false, err
		}
		if pgIsExpired(expiresAt) {
			if _, err := tx.Exec(`DELETE FROM `+st.table()+` WHERE "sid" = $1`, sid); err != nil {
				return false, errors.New(ErrSqlDeleteQuery)
			}
			return true, errors.New(ErrSessionExpired)
		}
		if _, err := tx.Exec(`UPDATE `+st.table()+` SET "sid" = $1 WHERE "sid" = $2`, newSid, sid); err != nil {
			return false, errors.New(ErrSqlUpdateQuery)
		}
		return false, nil
	})
	if err != nil {
		return "", err
	}
	return newSid, nil
}
