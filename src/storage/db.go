package storage

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
)

type DBClient struct {
	pool *pgxpool.Pool
	once sync.Once
}

var dbClient *DBClient

func InitDBClient() error {
	var err error
	dbClient = &DBClient{}
	dbClient.once.Do(func() {
		err = dbClient.createPool()
	})
	return err
}

func (dbc *DBClient) createPool() error {
	cfg := common.Cfg

	poolConfig, err := pgxpool.ParseConfig(cfg.PGDSN)
	if err != nil {
		return err
	}

	poolConfig.MaxConns = int32(cfg.PGMaxPoolSize)

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return err
	}

	dbc.pool = pool
	return nil
}

func GetDB() *DBClient {
	return dbClient
}

func (dbc *DBClient) Close() {
	if dbc.pool != nil {
		dbc.pool.Close()
	}
}

func (dbc *DBClient) GetSessions() ([]*models.Session, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dbc.pool.Acquire(ctx)
	if err != nil {
		log.Println("Cannot acquire the conn of db pool,err:", err)
		return nil, err

	}

	var session models.Session
	var sessionList []*models.Session

	rows, err := conn.Query(ctx, "select * from sessions;")
	if err != nil {
		log.Println("Cannot get the kernels", err)
		return nil, err
	}

	for rows.Next() {
		rows.Scan(&session.ID, &session.KernelInfo)

		sessionList = append(sessionList, &session)
	}

	return sessionList, nil
}

func (dbc *DBClient) GetSessionByID(id string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dbc.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	var session models.Session

	err = conn.QueryRow(ctx, "select * from sessions where id=$1;", id).Scan(&session.ID, &session.KernelInfo)
	if err != nil {
		log.Println("Cannot get the kernel "+id, err)
		return nil, err
	}
	return &session, nil
}

func (dbc *DBClient) DeleteSessionByIDS(ids []string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dbc.pool.Acquire(ctx)
	if err != nil {
		return -1, err
	}

	placeholders := make([]string, len(ids))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("DELETE FROM sessions WHERE id IN (%s)", strings.Join(placeholders, ","))

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		log.Println("Failed to delete kernels: ", err)
		return -1, err
	}
	return 0, nil

}

func (dbc *DBClient) SaveSession(id string, sessionJson []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := dbc.pool.Acquire(ctx)
	if err != nil {
		return err
	}

	kernelInfoJSON := pgtype.JSONB{
		Bytes:  sessionJson,
		Status: pgtype.Present,
	}

	_, err = conn.Exec(ctx, "INSERT INTO sessions (id, kernel_info) VALUES ($1, $2);", id, kernelInfoJSON)
	if err != nil {
		log.Println("Failed to insert kernel: ", err)
		return err
	}
	return nil

}
