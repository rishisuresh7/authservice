package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"authservice/models"
)

type RedisQueryer interface {
	GetString(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, timeOut time.Duration) error
	GetBytes(ctx context.Context, key string) ([]byte, error)
	GetDelString(ctx context.Context, key string) (string, error)
	IsRedisNil(err error) bool
	PushToChannel(ctx context.Context, notification *models.ChannelMessage) error
}

type redisQueryer struct {
	client *redis.Client
}

const listenerChannel = "notification-listener"

func NewRedisQueryer(d *redis.Client) RedisQueryer {
	return &redisQueryer{
		client: d,
	}
}

func (r *redisQueryer) Set(ctx context.Context, key string, value interface{}, timeOut time.Duration) error {
	res := r.client.Set(ctx, key, value, timeOut)
	if err := res.Err(); err != nil {
		return fmt.Errorf("set: unable to set key in redis: %s", err)
	}

	return nil
}

func (r *redisQueryer) GetString(ctx context.Context, key string) (string, error) {
	res := r.client.Get(ctx, key)
	if err := res.Err(); err != nil {
		return "", fmt.Errorf("getString: unable to get value from redis: %s", err)
	}

	return res.Val(), nil
}

func (r *redisQueryer) GetBytes(ctx context.Context, key string) ([]byte, error) {
	res := r.client.Get(ctx, key)
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("getBytes: unable to get value from redis: %s", err)
	}

	bytes, err := res.Bytes()
	if err != nil {
		return nil, fmt.Errorf("getBytes: unable to get bytes: %s", err)
	}

	return bytes, nil
}

func (r *redisQueryer) GetDelString(ctx context.Context, key string) (string, error) {
	res := r.client.GetDel(ctx, key)
	if err := res.Err(); err != nil {
		return "", fmt.Errorf("getDelString: unable to get value from redis: %s", err)
	}

	return res.Val(), nil
}

func (r *redisQueryer) IsRedisNil(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "redis: nil")
}

func (r *redisQueryer) PushToChannel(ctx context.Context, cm *models.ChannelMessage) error {
	bytes, _ := json.Marshal(cm)
	res := r.client.Publish(ctx, listenerChannel, bytes)
	if err := res.Err(); err != nil {
		return fmt.Errorf("pushToChannel: unable to push to channel: %s", err)
	}

	return nil
}