package logic

func sendCodeLuaTemplate() string {
	return `
-- 验证码key：code:biz:phone
local key = KEYS[1]
local cntKey = key..":cnt"
local sendCntKey = key..":cnt"

local code = ARGV[1]

-- 获得验证码的过期时间
local ttl = tonumber(redis.call("ttl", key))

-- -1 表示key存在但没有过期时间
if ttl == -1 then
    return -2
-- key不存在或者距离过期小于9分钟
elseif ttl == -2 or ttl < 540 then
    redis.call("set" ,key, code)
    redis.call("expire", key, 600)
    redis.call("set" ,cntKey, 3)
    redis.call("expire", cntKey, 600)
    return 0
else
    -- 距离上次发送验证码不超过1分钟
    return -1
end
`
}

func verifyCodeLuaScript() string {
	return `
-- 验证码key：code:biz:phone
local key = KEYS[1]
local cntKey = key..":cnt"

local code = ARGV[1]

-- 校验验证次数
local cnt = tonumber(redis.call("get",cntKey))
if cnt == nil then
	return -1
end
if cnt <= 0 then
	return -2
end

-- 判断验证码是否相等
local expectCode = redis.call("get",key)
if expectCode == code then
	redis.call("set",cntKey,-1)
	redis.call("expire", cntKey, 600)
	return 0
else
    redis.call("decr", cntKey, -1)
	return -3
end
`
}