package session

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlStore struct {
	db        *sql.DB
	tableName string
}

func (st *MysqlStore) CreateStore() string {
	return `CREATE TABLE "` + st.tableName + `" (
	"sid" varchar(64) NOT NULL,
	"data" mediumtext NOT NULL,
	"expires_at" datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
ALTER TABLE "` + st.tableName + `"
	ADD PRIMARY KEY ("sid");
COMMIT;`
}

func (st *MysqlStore) loadSessionData(sid string) (SessionData, error) {
	var data string
	var expiresAt string

	err := st.db.QueryRow("SELECT `data`, `expires_at` FROM `"+st.tableName+"` WHERE `sid` = ?", sid).Scan(&data, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(ErrSessionNotFound)
		}
		return nil, errors.New(ErrSqlSelectQuery)
	}

	parsedTime, err := time.Parse(dateFormat, expiresAt)
	if err != nil {
		return nil, errors.New(ErrParsingExpiresAt)
		// oppure lo setta zero parsedTime = time.Time{}
	}
	if time.Now().After(parsedTime) && !parsedTime.IsZero() {
		_ = st.Delete(sid)
		return nil, errors.New(ErrSessionExpired)
	}

	var parsedData SessionData
	err = json.Unmarshal([]byte(data), &parsedData)
	if err != nil {
		return nil, errors.New(ErrUnableUnserializeJson)
	}
	return parsedData, nil
}

func (st *MysqlStore) loadSessionRaw(sid string) (string, string, error) {
	var data string
	var expiresAt string

	err := st.db.QueryRow("SELECT `data`, `expires_at` FROM `"+st.tableName+"` WHERE `sid` = ?", sid).Scan(&data, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New(ErrSessionNotFound)
		}
		return "", "", errors.New(ErrSqlSelectQuery)
	}

	parsedTime, err := time.Parse(dateFormat, expiresAt)
	if err != nil {
		return "", "", errors.New(ErrParsingExpiresAt)
	}
	if time.Now().After(parsedTime) && !parsedTime.IsZero() {
		_ = st.Delete(sid)
		return "", "", errors.New(ErrSessionExpired)
	}

	return data, expiresAt, nil
}

func (st *MysqlStore) saveSessionData(sid string, parsedData SessionData) error {
	jsonData, err := json.Marshal(parsedData)
	if err != nil {
		return errors.New(ErrUnableSerializeJson)
	}

	_, err = st.db.Exec("UPDATE `"+st.tableName+"` SET `data` = ? WHERE `sid` = ?", string(jsonData), sid)
	if err != nil {
		return errors.New(ErrSqlUpdateQuery)
	}
	return nil
}

// Inizializza lo store.
func (st *MysqlStore) InitStore(db interface{}, tableName string) error {
	if !validTableName.MatchString(tableName) {
		return errors.New(ErrInvTableName)
	}
	st.db = db.(*sql.DB)
	st.tableName = tableName
	return nil
}

// Inizializza una sessione, restituendo il sid.
func (st *MysqlStore) Init(tm time.Time) (string, error) {
	sid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	data := SessionData{}
	expAt := tm.Format(dateFormat) // Converte in stringa
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", errors.New(ErrUnableSerializeJson)
	}

	tx, err := st.db.Begin()
	if err != nil {
		return "", err
	}

	committed := false
	defer func() {
		if !committed {
			// tx.Rollback()
			err := tx.Rollback()
			if err != nil {
				log.Println("[DEBUG] Error while tx.Rollback():", err)
			}
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO `" + st.tableName + "`(`sid`, `data`, `expires_at`) VALUES(?, ?, ?)")
	if err != nil {
		return "", errors.New(ErrSqlInsertQuery)
	}
	// defer stmt.Close()
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("[DEBUG] Error deferring stmt.Close():", err)
		}
	}()

	_, err = stmt.Exec(sid, string(jsonData), expAt)
	if err != nil {
		return "", errors.New(ErrSqlInsertQuery)
	}

	err = tx.Commit()
	if err != nil {
		return "", errors.New(ErrSqlInsertQuery)
	}
	committed = true

	return sid, nil
}

// Controlla se una sessione esiste (dato il sid).
func (st *MysqlStore) Check(sid string) bool {
	expiresAt := ""
	query := "SELECT `expires_at` FROM `" + st.tableName + "` WHERE `sid` = ?"
	err := st.db.QueryRow(query, sid).Scan(&expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			return false
		}
	} else {
		// expiresAt è stringa datetime di MySQL
		// Converti la stringa in un oggetto time.Time
		parsedTime, err := time.Parse(dateFormat, expiresAt)
		if err != nil || time.Now().After(parsedTime) && !parsedTime.IsZero() {
			_ = st.Delete(sid)
			return false
			// oppure lo setta zero parsedTime = time.Time{}
		} else {
			return true
		}
	}
}

// Setta il valore di una chiave di una sessione (dati sid, key e value).
func (st *MysqlStore) Set(sid string, key string, value interface{}) error {
	parsedData, err := st.loadSessionData(sid)
	if err != nil {
		return err
	}

	parsedData[key] = value
	return st.saveSessionData(sid, parsedData)
}

// Ottiene il valore di una chiave di una sessione (dati sid e key).
func (st *MysqlStore) Get(sid string, key string) (interface{}, error) {
	parsedData, err := st.loadSessionData(sid)
	if err != nil {
		return nil, err
	}

	val, ok := parsedData[key]
	if !ok {
		return nil, errors.New(ErrKeyNotFound)
	}

	return val, nil
}

// Cancella il valore collegato a una chiave di una sessione (dati sid e key).
func (st *MysqlStore) RemoveKey(sid string, key string) error {
	parsedData, err := st.loadSessionData(sid)
	if err != nil {
		return err
	}

	if _, ok := parsedData[key]; !ok {
		return errors.New(ErrKeyNotFound)
	}

	delete(parsedData, key)
	return st.saveSessionData(sid, parsedData)
}

// Cancella un'intera sessione (dato il sid).
func (st *MysqlStore) Delete(sid string) error {
	_, err := st.db.Exec("DELETE FROM `"+st.tableName+"` WHERE `sid` = ?", sid)
	if err != nil {
		return errors.New(ErrSqlDeleteQuery)
	}
	return nil
}

// Duplica la sessione con il sid dato (che viene eliminata).
func (st *MysqlStore) Refresh(sid string) (string, error) {
	parsedData, expiresAt, err := st.loadSessionRaw(sid)
	if err != nil {
		return "", err
	}

	newSid, err := generateSessionID()
	if err != nil {
		return "", err
	}

	tx, err := st.db.Begin()
	if err != nil {
		return "", err
	}

	committed := false
	defer func() {
		if !committed {
			// tx.Rollback()
			err := tx.Rollback()
			if err != nil {
				log.Println("[DEBUG] Error while tx.Rollback():", err)
			}
		}
	}()

	_, err = tx.Exec("DELETE FROM `"+st.tableName+"` WHERE `sid` = ?", sid)
	if err != nil {
		return "", err
	}

	stmt, err := tx.Prepare("INSERT INTO `" + st.tableName + "`(`sid`, `data`, `expires_at`) VALUES(?, ?, ?)")
	if err != nil {
		return "", errors.New(ErrSqlInsertQuery)
	}
	// defer stmt.Close()
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("[DEBUG] Error deferring stmt.Close():", err)
		}
	}()

	_, err = stmt.Exec(newSid, parsedData, expiresAt)
	if err != nil {
		return "", errors.New(ErrSqlInsertQuery)
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}
	committed = true

	return newSid, nil
}
