// storage/redis.go
package storage

import (
	"github.com/go-redis/redis"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(redisHost string, redisPort string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "",
		DB:       0,
	})

	return &RedisClient{Client: client}
}

func (r *RedisClient) Set(key string, value string) error {
	return r.Client.Set(key, value, 0).Err()
}

func (r *RedisClient) Get(key string) (string, error) {
	return r.Client.Get(key).Result()
}

func (r *RedisClient) LPush(key string, value string) error {
	return r.Client.LPush(key, value).Err()
}

func (r *RedisClient) RPop(key string) (string, error) {
	return r.Client.RPop(key).Result()
}

func (r *RedisClient) LRange(key string, start int64, end int64) ([]string, error) {
	return r.Client.LRange(key, start, end).Result()

}
