package idempotency

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type Store struct {
    client *redis.Client
    ttl    time.Duration
}

func NewStore(client *redis.Client, ttl time.Duration) *Store {
    return &Store{client: client, ttl: ttl}
}

func (s *Store) Seen(ctx context.Context, id string) (bool, error) {
    set, err := s.client.SetNX(ctx, "idempotency:"+id, 1, s.ttl).Result()
    if err != nil { return false, err }
    return !set, nil
}

func (s *Store) Forget(ctx context.Context, id string) error {
	return s.client.Del(ctx, "idempotency:"+id).Err()
}
