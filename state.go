package udpp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"
)

var (
	ping = []byte{0xEE, 0xFF, 0xFF, 0xEE}
	rdb  *redis.Client
)

func Setup(uri string) (err error) {
	if uri == "" {
		return fmt.Errorf("uri is empty")
	}
	rdb, err = NewRedis(uri)
	return err
}

func NewRedis(uri string) (*redis.Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	opt := &redis.Options{
		Addr:     u.Host,
		Username: u.User.Username(),
	}
	if pwd, ok := u.User.Password(); ok {
		opt.Password = pwd
	}
	c := redis.NewClient(opt)
	if err = c.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return c, nil
}
