local counter = 0
local thread_id = 0

-- 随机生成字符串的函数
function random_string(length)
    local chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    local result = ''
    for i = 1, length do
        local rand = math.random(#chars)
        result = result .. chars:sub(rand, rand)
    end
    return result
end

-- 初始化每个线程的设置
function setup(thread)
    thread:set("id", thread_id)
    thread_id = thread_id + 1
end

-- 定义每个线程执行的操作
function init(args)
    math.randomseed(os.time() + id)
end

-- 定义请求生成逻辑
function request()
    -- 每个请求生成随机的 email 和 password
    local email = random_string(8) .. "@example.com"
    local password = random_string(12)

    -- 请求体的 JSON 数据
    local body = string.format('{"email": "%s", "password": "%s"}', email, password)

    -- 构造 HTTP POST 请求
    local headers = {}
    headers["Content-Type"] = "application/json"

    -- 构建请求，指向注册 API 地址
    return wrk.format("POST", "/v1/user/sign", headers, body)
end

-- 设置请求速率
function done(summary, latency, requests)
    -- 打印执行情况
    print("Completed requests: ", summary.requests)
    print("Failed requests: ", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.status)
end
