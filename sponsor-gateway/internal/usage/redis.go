package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var incrementFixedWindow = redis.NewScript(`local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return count`)

type RedisFixedWindow struct {
	client redis.UniversalClient
	limit  int
	window time.Duration
}

func NewRedisFixedWindow(client redis.UniversalClient, limit int, window time.Duration) *RedisFixedWindow {
	return &RedisFixedWindow{client: client, limit: limit, window: window}
}
func (l *RedisFixedWindow) Allow(subject string, now time.Time) bool {
	if l == nil || l.limit <= 0 {
		return true
	}
	bucket := now.UTC().UnixNano() / l.window.Nanoseconds()
	key := fmt.Sprintf("sponsor:rate:%s:%d", subject, bucket)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	count, err := incrementFixedWindow.Run(ctx, l.client, []string{key}, l.window.Milliseconds()).Int64()
	if err != nil {
		return false
	}
	return count <= int64(l.limit)
}
func (l *RedisFixedWindow) Ping(ctx context.Context) error { return l.client.Ping(ctx).Err() }
