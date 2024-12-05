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