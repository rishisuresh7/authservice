package factory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	pg "github.com/jackc/pgx/v4/pgxpool"
)

var pgSync, redisSync sync.Once

func (f *factory) pgDriver() (*pg.Pool, error) {
	var err error
	pgSync.Do(func() {
		pgConfig, confErr := pg.ParseConfig(fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?search_path=%s",
			f.config.PgConfig.Username,
			f.config.PgConfig.Password,
			f.config.PgConfig.Host,
			f.config.PgConfig.Port,
			f.config.PgConfig.Database,
			f.config.PgConfig.Database,
		))
		if confErr != nil {
			err = confErr
			return
		}

		pgConfig.MaxConns = 100
		db, connErr := pg.ConnectConfig(context.TODO(), pgConfig)
		if connErr != nil {
			err = connErr
			return
		}

		f.pgConn = db
	})

	return f.pgConn, err
}

func (f *factory) redisDriver() (*redis.Client, error) {
	var err error
	redisSync.Do(func() {
		rdb := redis.NewClient(&redis.Options{
			Addr:        fmt.Sprintf("%s:%d", f.config.RedisConfig.Host, f.config.RedisConfig.Port),
			Username:    f.config.RedisConfig.Username,
			DB:          f.config.RedisConfig.Database,
			DialTimeout: 1 * time.Minute,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		res := rdb.Ping(ctx)
		if res.Err() != nil {
			err = res.Err()
			return
		}

		f.redisConn = rdb
	})

	return f.redisConn, err
}
