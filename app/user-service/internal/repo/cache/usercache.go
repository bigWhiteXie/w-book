package cache

import (
	"codexie.com/w-book-user/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var ErrKeyNotExist = errors.New("id not exist in cache")

type UserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) *UserCache {
	return &UserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

// Get
// @Description: 从redis中取出用户信息，缓存15分钟
// @param ctx
// @param id 用户id
// @return *model.User
// @return error redis的网络异常 | json反序列化异常 | key不存在
func (cache *UserCache) Get(ctx context.Context, id string) (*model.User, error) {
	key := cache.key(id)
	data, err := cache.cmd.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if data == "" {

	}
	user := &model.User{}
	return user, json.Unmarshal([]byte(data), user)
}

func (cache *UserCache) Set(ctx context.Context, user *model.User) error {
	key := cache.key(strconv.Itoa(user.Id))
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return cache.cmd.Set(ctx, key, data, cache.expiration).Err()
}

func (cache *UserCache) key(id string) string {
	return fmt.Sprintf("user:info:%s", id)
}
