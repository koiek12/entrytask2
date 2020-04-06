package cache

import (
	"git.garena.com/youngiek.song/entry_task/pkg/message"
	"github.com/go-redis/redis/v7"
)

type UserCache struct {
	client *redis.Client
}

func NewUserCache() *UserCache {
	return &UserCache{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
	}
}

func (c *UserCache) DelUserInfo(id string) error {
	res := c.client.Del(id)
	_, err := res.Result()
	return err
}

func (c *UserCache) SetUserInfo(user *message.User) error {
	res := c.client.HSet(user.Id, []string{"nickname", user.Nickname, "pic_path", user.PicPath})
	_, err := res.Result()
	return err
}

func (c *UserCache) GetUserInfo(id string) (*message.User, error) {
	res := c.client.HMGet(id, "nickname", "pic_path")
	vals, err := res.Result()
	if err != nil {
		return nil, err
	}
	if vals[0] == nil {
		return nil, nil
	}
	return &message.User{
		Id:       id,
		Nickname: vals[0].(string),
		PicPath:  vals[1].(string),
	}, nil
}
