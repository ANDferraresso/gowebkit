package session

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore salva le sessioni in Redis, una chiave (prefix+sid) per sessione,
// con la scadenza gestita dal TTL della chiave. Richiede Redis >= 6.0 (KEEPTTL).
type RedisStore struct {
	client *redis.Client
	ctx    context.Context
	prefix string
}

// maxTxRetries è il numero di tentativi di Set/RemoveKey quando la sessione
// viene modificata in contemporanea da un'altra richiesta.
const maxTxRetries = 10

func (st *RedisStore) CreateStore() string {
	return "" // Redis non richiede la memorizzazione in tabelle.
}

// InitStore inizializza lo store. client deve essere un *redis.Client.
func (st *RedisStore) InitStore(client interface{}, prefix string) error {
	c, ok := client.(*redis.Client)
	if !ok || c == nil {
		return errors.New(ErrInvStoreClient)
	}
	st.ctx = context.Background()
	st.client = c
	st.prefix = prefix
	return nil
}

// Init crea una sessione che scade a tm e ne restituisce il sid.
// Con tm zero la sessione non scade; con tm nel passato ritorna ErrInvExpiration.
func (st *RedisStore) Init(tm time.Time) (string, error) {
	var ttl time.Duration
	if !tm.IsZero() {
		ttl = time.Until(tm)
		if ttl <= 0 {
			return "", errors.New(ErrInvExpiration)
		}
	}
	sid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	if err := st.client.Set(st.ctx, st.prefix+sid, "{}", ttl).Err(); err != nil {
		return "", err
	}
	return sid, nil
}

func (st *RedisStore) Check(sid string) bool {
	exists, err := st.client.Exists(st.ctx, st.prefix+sid).Result()
	return err == nil && exists == 1
}

func (st *RedisStore) Get(sid string, key string) (interface{}, error) {
	raw, err := st.client.Get(st.ctx, st.prefix+sid).Result()
	if err == redis.Nil {
		return nil, errors.New(ErrSessionNotFound)
	} else if err != nil {
		return nil, err
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

// modify legge la sessione, applica fn e la riscrive in modo atomico
// (WATCH/MULTI/EXEC), mantenendo il TTL della chiave. Se un'altra richiesta
// modifica o fa scadere la sessione nel frattempo, l'operazione viene ritentata.
func (st *RedisStore) modify(sid string, fn func(SessionData) error) error {
	key := st.prefix + sid
	txf := func(tx *redis.Tx) error {
		raw, err := tx.Get(st.ctx, key).Result()
		if err == redis.Nil {
			return errors.New(ErrSessionNotFound)
		} else if err != nil {
			return err
		}
		var data SessionData
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			return errors.New(ErrUnableUnserializeJson)
		}
		if data == nil {
			data = SessionData{}
		}
		if err := fn(data); err != nil {
			return err
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return errors.New(ErrUnableSerializeJson)
		}
		_, err = tx.TxPipelined(st.ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(st.ctx, key, jsonData, redis.KeepTTL)
			return nil
		})
		return err
	}

	for i := 0; i < maxTxRetries; i++ {
		err := st.client.Watch(st.ctx, txf, key)
		if !errors.Is(err, redis.TxFailedErr) {
			return err
		}
	}
	return errors.New(ErrSessionConflict)
}

func (st *RedisStore) Set(sid string, key string, value interface{}) error {
	return st.modify(sid, func(data SessionData) error {
		data[key] = value
		return nil
	})
}

func (st *RedisStore) RemoveKey(sid string, key string) error {
	return st.modify(sid, func(data SessionData) error {
		if _, ok := data[key]; !ok {
			return errors.New(ErrKeyNotFound)
		}
		delete(data, key)
		return nil
	})
}

func (st *RedisStore) Delete(sid string) error {
	return st.client.Del(st.ctx, st.prefix+sid).Err()
}

// Refresh sostituisce il sid della sessione con uno nuovo, mantenendo dati e
// TTL. RENAME è atomico e conserva la scadenza della chiave.
func (st *RedisStore) Refresh(sid string) (string, error) {
	newSid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	err = st.client.Rename(st.ctx, st.prefix+sid, st.prefix+newSid).Err()
	if err != nil {
		if strings.Contains(err.Error(), "no such key") {
			return "", errors.New(ErrSessionNotFound)
		}
		return "", err
	}
	return newSid, nil
}
