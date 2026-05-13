package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/Fipaan/ap2-uni/order-service/internal/domain"
    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
    ttl     time.Duration
}

func NewRedisCache(addr string, ttl time.Duration) (*RedisCache, error) {
    c := redis.NewClient(&redis.Options{Addr: addr})
    if err := c.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("redis ping: %w", err)
    }
    return &RedisCache{client: c, ttl: ttl}, nil
}

func key(id string) string { return "order:" + id }

func (r *RedisCache) Get(ctx context.Context, id string) (*domain.Order, error) {
    val, err := r.client.Get(ctx, key(id)).Bytes()
    if err != nil { return nil, err }
    var o domain.Order
    return &o, json.Unmarshal(val, &o)
}

func (r *RedisCache) Set(ctx context.Context, o *domain.Order) error {
    b, err := json.Marshal(o)
    if err != nil { return err }
    return r.client.Set(ctx, key(o.ID), b, r.ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, id string) error {
    return r.client.Del(ctx, key(id)).Err()
}

func (r *RedisCache) Close() error {
    return r.client.Close()
}
