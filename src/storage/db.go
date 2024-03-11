package storage

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"zjuici.com/tablegpt/jkpmanager/src/models"
)

// DBClient 数据库客户端
type DBClient struct {
	pool *pgxpool.Pool
	once sync.Once
}

var dbClient *DBClient

// InitDBClient 初始化数据库客户端(单例模式)
func InitDBClient(cfg *models.Config) error {
	var err error
	dbClient = &DBClient{}
	dbClient.once.Do(func() {
		err = dbClient.createPool(cfg)
	})
	return err
}

// createPool 创建数据库连接池
func (dbc *DBClient) createPool(cfg *models.Config) error {
	poolConfig, err := pgxpool.ParseConfig(cfg.PGDns)
	if err != nil {
		return err
	}

	// 设置连接池选项
	poolConfig.MaxConns = int32(cfg.PGMaxPoolSize)

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return err
	}

	dbc.pool = pool
	return nil
}

// GetDB 获取数据库连接池实例
func GetDB() *pgxpool.Pool {
	return dbClient.pool
}

// Close 关闭数据库连接池
func (dbc *DBClient) Close() {
	if dbc.pool != nil {
		dbc.pool.Close()
	}
}
