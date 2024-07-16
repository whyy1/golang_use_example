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
	//TODO 动态修改限速器参数
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
