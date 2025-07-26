package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client     *redis.Client
	ctx        context.Context
	prefix     string
	expiration time.Duration
}

func (st *RedisStore) CreateStore() string {
	return "" // Redis non richiede la memorizzazione in tabelle.
}

func (st *RedisStore) InitStore(client interface{}, prefix string) error {
	st.ctx = context.Background()
	st.client = client.(*redis.Client)
	st.prefix = prefix
	st.expiration = 24 * time.Hour // Durata predefinita della sessione.
	return nil
}

func (st *RedisStore) Init(tm time.Time) (string, error) {
	sid, err := generateSessionID()
	if err != nil {
		return "", err
	}
	data := SessionData{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", errors.New(ErrUnableSerializeJson)
	}
	ttl := time.Until(tm)
	err = st.client.Set(st.ctx, st.prefix+sid, jsonData, ttl).Err()
	if err != nil {
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

func (st *RedisStore) Set(sid string, key string, value interface{}) error {
	raw, err := st.client.Get(st.ctx, st.prefix+sid).Result()

	if err == redis.Nil {
		return errors.New(ErrSessionNotFound)
	} else if err != nil {
		return err
	}

	var data SessionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return errors.New(ErrUnableUnserializeJson)
	}

	data[key] = value
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New(ErrUnableSerializeJson)
	}

	ttl, _ := st.client.TTL(st.ctx, st.prefix+sid).Result()

	return st.client.Set(st.ctx, st.prefix+sid, jsonData, ttl).Err()
}

func (st *RedisStore) RemoveKey(sid string, key string) error {
	raw, err := st.client.Get(st.ctx, st.prefix+sid).Result()
	if err == redis.Nil {
		return errors.New(ErrSessionNotFound)
	} else if err != nil {
		return err
	}

	var data SessionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return errors.New(ErrUnableUnserializeJson)
	}

	if _, ok := data[key]; !ok {
		return errors.New(ErrKeyNotFound)
	}
	delete(data, key)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New(ErrUnableSerializeJson)
	}

	ttl, _ := st.client.TTL(st.ctx, st.prefix+sid).Result()
	return st.client.Set(st.ctx, st.prefix+sid, jsonData, ttl).Err()
}

func (st *RedisStore) Delete(sid string) error {
	return st.client.Del(st.ctx, st.prefix+sid).Err()
}

func (st *RedisStore) Refresh(sid string) (string, error) {
	raw, err := st.client.Get(st.ctx, st.prefix+sid).Result()
	if err == redis.Nil {
		return "", errors.New(ErrSessionNotFound)
	} else if err != nil {
		return "", err
	}

	ttl, _ := st.client.TTL(st.ctx, st.prefix+sid).Result()
	newSid, err := generateSessionID()
	if err != nil {
		return "", err
	}

	err = st.client.Set(st.ctx, st.prefix+newSid, raw, ttl).Err()
	if err != nil {
		return "", err
	}

	err = st.client.Del(st.ctx, st.prefix+sid).Err()
	if err != nil {
		return "", err
	}

	return newSid, nil
}

/*
func NewRedisClient() *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", // Indirizzo del server Redis
        Password: "", // Nessuna password per default
        DB:       0,  // Database di default
    })
    return rdb
}

func CreateSession(session *Session) error {
    rdb := NewRedisClient()
    err := rapp.Db.Set(ctx, session.ID, session.Userdata, 24*time.Hour).Err()
    if err != nil {
        return err
    }
    return nil
}


func GetSession(sessionID string) (*Session, error) {
    rdb := NewRedisClient()
    val, err := rapp.Db.Get(ctx, sessionID).Result()
    if err != nil {
        return nil, err
    }
    // Assumendo che i dati dell'utente siano memorizzati come stringa,
    // qui dovresti deserializzare val in un oggetto Session.
    // L'esempio semplifica questa parte, ma nella pratica dovresti
    // convertire val in una struttura dati appropriata.
    return &Session{ID: sessionID, Userdata: val}, nil
}


*/
