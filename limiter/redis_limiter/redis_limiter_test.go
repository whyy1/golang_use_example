package redis_limiter

import (
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"os"
	"sync"
	"testing"
	"time"
)

func TestRedisLimiter_RedisWaitAllowN(t *testing.T) {
	var globalwg sync.WaitGroup
	globalwg.Add(2)
	go RunTest("test1", &globalwg)
	go RunTest("test2", &globalwg)
	globalwg.Wait()
}

func RunTest(key string, wg1 *sync.WaitGroup) {
	defer wg1.Done()
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("Redis_Addr"),
		Password: os.Getenv("Redis_Pwd"),
		DB:       0,
	})
	limiters := RedisLimiter{Limiter: *redis_rate.NewLimiter(rdb)}
	chatBots := []string{"bot1"}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		for _, chatBot := range chatBots {
			wg.Add(1)
			go func(chatBot string, i int) {
				defer wg.Done()
				//var err error
				startTime := time.Now()
				withTimeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelFunc()
				ok, err := limiters.RedisWaitAllowN(withTimeout, chatBot, redis_rate.PerSecond(3), 2)
				if err != nil {
					fmt.Printf("%d限流器等待获取令牌失败, chatBotName=%s,等待时间%v error=%v\n", i, chatBot, time.Since(startTime), err)
					return
				}
				if ok {
					fmt.Printf("%v成功时间%v\n", key, time.Now())
				}
				//fmt.Printf("%schatBotName=%s 限流器%d等待时间%v等待结果%v, \n", key, chatBot, i, time.Since(startTime), ok)

			}(chatBot, i)
			//time.Sleep(10 * time.Millisecond) // 增加请求间隔，避免过多请求立即消耗掉令牌
		}
	}

	wg.Wait() // 等待所有 goroutine 完成
}
