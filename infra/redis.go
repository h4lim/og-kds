package infra

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisDB IRedisConfig

const (
	JsonRedis  = "1"
	XmlType    = "2"
	SingleType = "3"
)

type RedisModel struct {
	Domain         string
	Port           string
	Password       string
	SecondDuration int
}

type IRedisConfig interface {
	Open() *error
	Set(key string, redisType string, value any, duration int) *error
	Get(key string, redisType string) (any, *error)
	Delete(key string) *error
}

func InitRedis(model RedisModel) {
	RedisDB = NewRedisConfig(model)
}

func NewRedisConfig(model RedisModel) IRedisConfig {
	return RedisModel{
		Domain:   model.Domain,
		Port:     model.Port,
		Password: model.Password,
	}
}

func (r RedisModel) Open() *error {

	if _, err := open(r); err != nil {
		return err
	}

	return nil
}

func (r RedisModel) Set(key string, redisType string, value any, duration int) *error {

	client, err := open(r)
	if err != nil {
		return err
	}
	defer client.Close()

	_duration := time.Duration(duration) * time.Second
	switch redisType {
	case JsonRedis:
		jsonResult, err := json.Marshal(&value)
		if err != nil {
			return &err
		}

		if cmdStatus := client.Set(context.Background(), key, string(jsonResult), _duration); cmdStatus != nil {
			newError := errors.New(cmdStatus.String())
			return &newError
		}

		return nil
	case XmlType:
		jsonResult, err := xml.Marshal(&value)
		if err != nil {
			return &err
		}

		if cmdStatus := client.Set(context.Background(), key, string(jsonResult), _duration); cmdStatus != nil {
			newError := errors.New(cmdStatus.String())
			return &newError
		}

		return nil
	case SingleType:
		if cmdStatus := client.Set(context.Background(), key, value, _duration); cmdStatus != nil {
			newError := errors.New(cmdStatus.String())
			return &newError
		}

		return nil
	default:
		newError := errors.New("invalid redis type")
		return &newError
	}

}

func (r RedisModel) Get(key string, redisType string) (any, *error) {

	client, err := open(r)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	value, errGet := client.Get(context.Background(), key).Result()
	if errGet != nil {
		return nil, &errGet
	}

	switch redisType {
	case JsonRedis:
		var jsonResult interface{}
		if err := json.Unmarshal([]byte(value), &jsonResult); err != nil {
			return nil, &err
		}
	case XmlType:
		var xmlResult interface{}
		if err := xml.Unmarshal([]byte(value), &xmlResult); err != nil {
			return nil, &err
		}
	case SingleType:
		return value, nil
	}

	newError := errors.New("invalid redis type")
	return nil, &newError

}

func (r RedisModel) Delete(key string) *error {

	client, err := open(r)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Del(context.Background(), key).Err(); err != nil {
		return &err
	}

	return nil
}

func open(model RedisModel) (*redis.Client, *error) {

	client := redis.NewClient(&redis.Options{
		Addr:     model.Domain + ":" + model.Port,
		Password: model.Password,
		DB:       0,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, &err
	}

	return client, nil
}
