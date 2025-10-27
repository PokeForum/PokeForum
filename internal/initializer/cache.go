package initializer

import (
	"strconv"
	"time"

	"github.com/PokeForum/PokeForum/internal/configs"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

func Cache() *redis.Pool {
	m := configs.Config.Cache

	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				m.Host+":"+strconv.Itoa(m.Port),
				redis.DialDatabase(m.DB),
				redis.DialPassword(m.Password),
			)
			if err != nil {
				configs.Log.Panic("Failed to create Redis connection: ", zap.Error(err))
			}
			return c, nil
		},
	}
}
