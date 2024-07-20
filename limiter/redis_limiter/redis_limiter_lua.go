package redis_limiter

import "github.com/redis/go-redis/v9"

// redis_rate限流命令实现的脚本

// Copyright (c) 2017 Pavel Pravosud
// https://github.com/rwz/redis-gcra/blob/master/vendor/perform_gcra_ratelimit.lua
var allowN = redis.NewScript(`
-- this script has side-effects, so it requires replicate commands mode
redis.replicate_commands()
--
local rate_limit_key = KEYS[1] --Redis 中用于存储当前令牌桶状态的键。
local burst = ARGV[1] -- 允许的最大突发请求数。
local rate = ARGV[2] --令牌生成速率（每秒生成多少个令牌）。
local period = ARGV[3] --生成令牌的时间周期。
local cost = tonumber(ARGV[4]) --每次请求消耗的令牌数。

local emission_interval = period / rate --每个令牌生成所需的时间。
local increment = emission_interval * cost --此次请求所需的时间增量。
local burst_offset = emission_interval * burst --允许的突发时间偏移。

-- redis returns time as an array containing two integers: seconds of the epoch
-- time (10 digits) and microseconds (6 digits). for convenience we need to
-- convert them to a floating point number. the resulting number is 16 digits,
-- bordering on the limits of a 64-bit double-precision floating point number.
-- adjust the epoch to be relative to Jan 1, 2017 00:00:00 GMT to avoid floating
-- point problems. this approach is good until "now" is 2,483,228,799 (Wed, 09
-- Sep 2048 01:46:39 GMT), when the adjusted value is 16 digits.
--将当前时间转换为浮点数格式，基于 2017 年 1 月 1 日，避免浮点数精度问题
local jan_1_2017 = 1483228800
local now = redis.call("TIME")
now = (now[1] - jan_1_2017) + (now[2] / 1000000)

--获取当前令牌桶的时间戳
local tat = redis.call("GET", rate_limit_key)

if not tat then
--如果 tat 为空，将 tat 初始化为当前时间 now。
  tat = now
else
--如果 tat 存在，将其转换为数字类型。
  tat = tonumber(tat)
end

--计算 tat 和 now 的最大值，确保 tat 不早于当前时间。如果 tat 早于当前时间，说明之前的请求可能已经被处理过，因此需要更新为当前时间。
tat = math.max(tat, now)

--根据当前时间 tat 加上请求的时间增量 increment，计算新的时间戳 new_tat。
local new_tat = tat + increment
--计算允许请求的时间 allow_at，等于新的时间戳 new_tat 减去突发偏移量 burst_offset。
local allow_at = new_tat - burst_offset

--计算当前时间 now 与允许请求时间 allow_at 之间的差值 diff
local diff = now - allow_at
--根据时间差值 diff 除以生成令牌的时间间隔 emission_interval，计算出剩余的令牌数 remaining

local remaining = diff / emission_interval

--如果剩余令牌数小于 0，说明当前没有足够的令牌可以处理请求
if remaining < 0 then
  local reset_after = tat - now  --计算重置时间 reset_after，等于 tat 减去当前时间 now。
  local retry_after = diff * -1 --计算重试时间 retry_after，等于 diff 的相反数（因为 diff 是负数）。
  return {
    0, -- allowed   
    0, -- remaining
    tostring(retry_after),-- 重试时间
    tostring(reset_after),-- 重置时间
  }
end

--计算新的重置时间 reset_after，等于新的时间戳 new_tat 减去当前时间 now。
local reset_after = new_tat - now
if reset_after > 0 then --如果重置时间 reset_after 大于 0，更新 Redis 中的令牌桶状态
  redis.call("SET", rate_limit_key, new_tat, "EX", math.ceil(reset_after)) --SET 命令用于将 rate_limit_key 设置为新的时间戳 new_tat。EX 选项用于设置键的过期时间为 reset_after，向上取整。
end

local retry_after = -1 --设置重试时间为 -1（表示不需要重试）
return {cost, remaining, tostring(retry_after), tostring(reset_after)}
`)

var allowAtMost = redis.NewScript(`
-- this script has side-effects, so it requires replicate commands mode
redis.replicate_commands()

local rate_limit_key = KEYS[1]
local burst = ARGV[1]
local rate = ARGV[2]
local period = ARGV[3]
local cost = tonumber(ARGV[4])

local emission_interval = period / rate
local burst_offset = emission_interval * burst

-- redis returns time as an array containing two integers: seconds of the epoch
-- time (10 digits) and microseconds (6 digits). for convenience we need to
-- convert them to a floating point number. the resulting number is 16 digits,
-- bordering on the limits of a 64-bit double-precision floating point number.
-- adjust the epoch to be relative to Jan 1, 2017 00:00:00 GMT to avoid floating
-- point problems. this approach is good until "now" is 2,483,228,799 (Wed, 09
-- Sep 2048 01:46:39 GMT), when the adjusted value is 16 digits.
local jan_1_2017 = 1483228800
local now = redis.call("TIME")
now = (now[1] - jan_1_2017) + (now[2] / 1000000)

local tat = redis.call("GET", rate_limit_key)

if not tat then
  tat = now
else
  tat = tonumber(tat)
end

tat = math.max(tat, now)

local diff = now - (tat - burst_offset)
local remaining = diff / emission_interval

if remaining < 1 then
  local reset_after = tat - now
  local retry_after = emission_interval - diff
  return {
    0, -- allowed
    0, -- remaining
    tostring(retry_after),
    tostring(reset_after),
  }
end

if remaining < cost then
  cost = remaining
  remaining = 0
else
  remaining = remaining - cost
end

local increment = emission_interval * cost
local new_tat = tat + increment

local reset_after = new_tat - now
if reset_after > 0 then
  redis.call("SET", rate_limit_key, new_tat, "EX", math.ceil(reset_after))
end

return {
  cost,
  remaining,
  tostring(-1),
  tostring(reset_after),
}
`)
