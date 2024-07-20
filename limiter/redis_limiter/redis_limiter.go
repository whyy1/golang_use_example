package redis_limiter

import (
	"context"
	"github.com/go-redis/redis_rate/v10"
	"time"
)

type RedisLimiter struct {
	redis_rate.Limiter
}

func (limiter *RedisLimiter) RedisWaitAllowN(ctx context.Context, key string, limit redis_rate.Limit, n int) (bool, error) {
	//如果同一时刻对同一个key传入不同的速率，会按照各自速率计算，并推进令牌桶的时间戳tat
	res, err := limiter.AllowN(ctx, key, limit, n)
	if err != nil {
		return false, err
	}
	if res.Allowed > 0 {
		//成功返回
		return true, nil
	}

	for {
		select {
		case <-time.After(res.ResetAfter):
			res, err := limiter.AllowN(ctx, key, limit, n)
			if err != nil {
				return false, err
			}
			if res.Allowed > 0 {
				//成功返回
				return true, nil
			}
		case <-ctx.Done():
			//超时返回
			return false, ctx.Err()
		}
	}
}
