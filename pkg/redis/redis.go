package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eiam-platform/config"
	"eiam-platform/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RDB *redis.Client
var ctx = context.Background()

// InitRedis initialize Redis connection
func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Test connection
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis connection failed: %v", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)

	return nil
}

// GetRedis get Redis client
func GetRedis() *redis.Client {
	return RDB
}

// Close close Redis connection
func Close() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}

// Set set key-value pair
func Set(key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

// Get get value
func Get(key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// GetBytes get bytes array
func GetBytes(key string) ([]byte, error) {
	return RDB.Get(ctx, key).Bytes()
}

// Del delete key
func Del(keys ...string) error {
	return RDB.Del(ctx, keys...).Err()
}

// Exists check if key exists
func Exists(key string) (bool, error) {
	count, err := RDB.Exists(ctx, key).Result()
	return count > 0, err
}

// Expire set expiration time
func Expire(key string, expiration time.Duration) error {
	return RDB.Expire(ctx, key, expiration).Err()
}

// TTL get remaining expiration time
func TTL(key string) (time.Duration, error) {
	return RDB.TTL(ctx, key).Result()
}

// SetJSON set JSON object
func SetJSON(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Set(key, jsonData, expiration)
}

// GetJSON get JSON object
func GetJSON(key string, dest interface{}) error {
	data, err := GetBytes(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// HSet set hash field
func HSet(key string, field string, value interface{}) error {
	return RDB.HSet(ctx, key, field, value).Err()
}

// HGet get hash field
func HGet(key string, field string) (string, error) {
	return RDB.HGet(ctx, key, field).Result()
}

// HDel delete hash field
func HDel(key string, fields ...string) error {
	return RDB.HDel(ctx, key, fields...).Err()
}

// HExists check if hash field exists
func HExists(key string, field string) (bool, error) {
	return RDB.HExists(ctx, key, field).Result()
}

// HGetAll get all hash fields
func HGetAll(key string) (map[string]string, error) {
	return RDB.HGetAll(ctx, key).Result()
}

// Incr increment
func Incr(key string) (int64, error) {
	return RDB.Incr(ctx, key).Result()
}

// Decr decrement
func Decr(key string) (int64, error) {
	return RDB.Decr(ctx, key).Result()
}

// SAdd add to set
func SAdd(key string, members ...interface{}) error {
	return RDB.SAdd(ctx, key, members...).Err()
}

// SRem remove from set
func SRem(key string, members ...interface{}) error {
	return RDB.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func SMembers(key string) ([]string, error) {
	return RDB.SMembers(ctx, key).Result()
}

// SIsMember 检查是否为集合成员
func SIsMember(key string, member interface{}) (bool, error) {
	return RDB.SIsMember(ctx, key, member).Result()
}

// ZAdd 添加到有序集合
func ZAdd(key string, score float64, member interface{}) error {
	return RDB.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRem 从有序集合删除
func ZRem(key string, members ...interface{}) error {
	return RDB.ZRem(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围
func ZRange(key string, start, stop int64) ([]string, error) {
	return RDB.ZRange(ctx, key, start, stop).Result()
}

// ZScore 获取有序集合成员分数
func ZScore(key string, member string) (float64, error) {
	return RDB.ZScore(ctx, key, member).Result()
}

// Publish 发布消息
func Publish(channel string, message interface{}) error {
	return RDB.Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func Subscribe(channels ...string) *redis.PubSub {
	return RDB.Subscribe(ctx, channels...)
}

// HealthCheck Redis健康检查
func HealthCheck() error {
	_, err := RDB.Ping(ctx).Result()
	return err
}

// FlushDB 清空当前数据库
func FlushDB() error {
	return RDB.FlushDB(ctx).Err()
}

// Keys 获取匹配的键
func Keys(pattern string) ([]string, error) {
	return RDB.Keys(ctx, pattern).Result()
}
