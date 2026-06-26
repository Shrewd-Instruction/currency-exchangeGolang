package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
}

func newCacheService(addr, password string, db int) *CacheService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Warn().Msgf("redis not available: %v", err)
		return &CacheService{client: nil}
	}

	log.Info().Msgf("connected to redis at %s", addr)
	return &CacheService{client: rdb}
}

func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	if c == nil || c.client == nil {
		return "", fmt.Errorf("cache not available")
	}
	return c.client.Get(ctx, key).Result()
}

func (c *CacheService) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, key, val, ttl).Err()
}

func (c *CacheService) Ping() error {
	if c == nil || c.client == nil {
		return fmt.Errorf("redis not connected")
	}
	return c.client.Ping(context.Background()).Err()
}

func (c *CacheService) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
