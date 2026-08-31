package redis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb       *redis.Client
	isLive    bool
	inMemLock sync.Map
}

func NewClient(host, port, password string) *Client {
	addr := fmt.Sprintf("%s:%s", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client := &Client{rdb: rdb, isLive: false}

	if err := rdb.Ping(ctx).Err(); err == nil {
		client.isLive = true
		log.Printf("[Redis] Connected to Redis at %s", addr)
	} else {
		log.Printf("[Redis] Redis unavailable at %s (%v). Falling back to resilient in-memory store.", addr, err)
	}

	return client
}

// SetIdempotencyKey sets a lock key with TTL. Returns true if lock was acquired, false if key already exists.
func (c *Client) SetIdempotencyKey(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if c.isLive {
		ok, err := c.rdb.SetNX(ctx, "idemp:"+key, "locked", ttl).Result()
		if err == nil {
			return ok, nil
		}
	}

	// In-memory fallback
	_, loaded := c.inMemLock.LoadOrStore(key, time.Now().Add(ttl))
	if loaded {
		return false, nil // already locked
	}
	// Expire asynchronously
	go func() {
		time.Sleep(ttl)
		c.inMemLock.Delete(key)
	}()
	return true, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.isLive {
		return c.rdb.Set(ctx, key, value, ttl).Err()
	}
	c.inMemLock.Store(key, value)
	return nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c.isLive {
		return c.rdb.Get(ctx, key).Result()
	}
	val, ok := c.inMemLock.Load(key)
	if !ok {
		return "", fmt.Errorf("key not found")
	}
	return fmt.Sprintf("%v", val), nil
}

func (c *Client) Close() {
	if c.rdb != nil {
		_ = c.rdb.Close()
	}
}
