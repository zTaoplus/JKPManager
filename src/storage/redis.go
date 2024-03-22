// storage/redis.go
package storage

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
)

type RedisClient struct {
	Client *redis.Client
	once   sync.Once
}

var redisClient *RedisClient

func InitRedisClient() error {

	var err error

	redisClient = &RedisClient{}
	redisClient.once.Do(func() {
		err = redisClient.createRedisClient()
	})
	return err

}

func (r *RedisClient) createRedisClient() error {
	opts, err := redis.ParseURL(common.Cfg.RedisDSN)
	if err != nil {
		log.Println("Cannot parse url from cfg redis dsn")
		return err
	}
	opts.PoolFIFO = true
	opts.MinIdleConns = 5
	opts.IdleTimeout = 5 * time.Minute
	opts.MaxConnAge = 1 * time.Hour

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	healthy := false

	// retries 5 count to ping redis

	for i := 0; i < 5; i++ {
		if _, err := client.Ping(ctx).Result(); err != nil {
			log.Printf("Failed to ping redis, retry count: %v", i+1)
			time.Sleep(1500 * time.Millisecond)
			continue
		} else {
			healthy = true
			break
		}
	}

	if !healthy {
		return errors.New("cannot init redis client, maybe the server cannot ping")
	}
	r.Client = client

	return nil
}

func GetRedisClient() *RedisClient {
	return redisClient
}

func (r *RedisClient) GetSessions() ([]*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.Client.HGetAll(ctx, common.Cfg.KernelsSessionKey).Result()
	if err != nil {
		return nil, err
	}

	var sessionList []*models.Session

	for k, v := range res {
		session := &models.Session{ID: k, KernelInfo: []byte(v)}
		sessionList = append(sessionList, session)
	}

	return sessionList, nil
}

func (r *RedisClient) GetSessionByID(id string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.Client.HGet(ctx, common.Cfg.KernelsSessionKey, id).Result()
	if err != nil {
		return nil, err
	}

	return &models.Session{ID: id, KernelInfo: []byte(res)}, nil
}

func (r *RedisClient) DeleteSessionByIDS(ids []string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.Client.HDel(ctx, common.Cfg.KernelsSessionKey, ids...).Result()

}

func (r *RedisClient) SaveSession(id string, sessionJson []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.Client.HSet(ctx, common.Cfg.KernelsSessionKey, id, sessionJson).Err()

}
