package cache

func updateCntTemplate() string {
	return `
local key = KEYS[1]
local cntKey = ARGV[1]
local delta = tonumber(ARGV[2])
local exist = redis.call("EXISTS", key)

if exist == 1 then
    redis.call("HINCRBY", key, cntKey, delta)
    return 1
else
    return 0
end
`
}

func increResourceLike() string {
	return `
-- 参数说明
-- KEYS[1]: zset 的 key
-- ARGV[1]: member
-- ARGV[2]: 分数增量 incre (正数或负数)

-- 检查成员是否存在
local exists = redis.call("ZSCORE", KEYS[1], ARGV[1])
if exists then
    -- 成员存在，更新分数
    redis.call("ZINCRBY", KEYS[1], ARGV[2], ARGV[1])
end

-- 返回操作状态
return exists and "Updated" or "Not Found"
        
`
}

func updateTopLike() string {
	return `
-- 参数说明：
-- KEYS[1]: zset 的 key
-- ARGV: 偶数个参数，第一个是资源 ID，第二个是对应的点赞数，依此类推

-- 删除原有的 key
redis.call("DEL", KEYS[1])

-- 插入新的数据到 zset
for i = 1, #ARGV, 2 do
    local id = ARGV[i]
    local like_count = ARGV[i + 1]
    redis.call("ZADD", KEYS[1], like_count, id)
end

return "OK"
`
}
