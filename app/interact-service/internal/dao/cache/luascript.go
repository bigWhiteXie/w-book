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
fi
return 0
`
}
