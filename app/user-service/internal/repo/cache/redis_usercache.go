package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"codexie.com/w-book-common/common/codeerr"
	"codexie.com/w-book-user/internal/model"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

var ErrKeyNotExist = errors.New("id not exist in cache")

type RedisUserCache struct {
	// mockgen -package=redismocks -destination=mocks/repo/cache/cmd/cmd_mock.go github.com/redis/go-redis/v9 Cmdable
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(cmd *redis.Client) UserCache {
	return &RedisUserCache{
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
func (cache *RedisUserCache) Get(ctx context.Context, id string) (*model.User, error) {
	key := cache.key(id)
	data, err := cache.cmd.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.Wrap(codeerr.WithCode(codeerr.SystemErrCode, "[RedisUserCache_Get]redis中获取key=%s失败:%s", key, err), "")
	}

	if data == "" {
		logx.WithContext(ctx).Infof("[RedisUserCache_Get]redis中不存在该key=%s", key)
		return nil, nil
	}
	user := &model.User{}
	return user, json.Unmarshal([]byte(data), user)
}

func (cache *RedisUserCache) Set(ctx context.Context, user *model.User) error {
	key := cache.key(strconv.Itoa(user.Id))
	data, err := json.Marshal(user)
	if err != nil {
		return errors.Wrap(codeerr.WithCode(codeerr.SystemErrCode, "[RedisUserCache_Set] 序列化用户信息%v失败:%s", user, err), "")
	}
	if err := cache.cmd.Set(ctx, key, data, cache.expiration).Err(); err != nil {
		return errors.Wrap(codeerr.WithCode(codeerr.SystemErrCode, "[RedisUserCache_Set] 写入用户信息%v到缓存失败:%s", user, err), "")
	}

	return nil
}

func (cache *RedisUserCache) key(id string) string {
	return fmt.Sprintf("user:info:%s", id)
}
