package session

// ValkeyStore salva le sessioni in un server Valkey standalone usando
// go-redis/v9. InitStore riceve un *redis.Client connesso a Valkey.
// Condivide con RedisStore il formato delle chiavi, il JSON e la gestione del TTL.
type ValkeyStore struct {
	RedisStore
}

var _ Store = (*ValkeyStore)(nil)
