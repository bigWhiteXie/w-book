-- 基础 URL 和 token 变量
local token = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjk2OTkyNjAsIlVpZCI6MSwiU3NpZCI6ImQ0Y2I4NmVmLTkwOGUtMTFlZi04NTQ1LTAwMGMyOWU1Yjk3OSIsIlVzZXJBZ3JlbnQiOiJBcGlmb3gvMS4wLjAgKGh0dHBzOi8vYXBpZm94LmNvbSkifQ.1DPlqLmrXWG0xONW1EKM6JltWfKlU3uQsoZsZGl-LkckzN8ACOmzs402YZz-v5XxzLguOZP6cRk6FJ7VU3Shew"  -- 从环境变量获取 token

-- 如果没有获取到 token，则抛出错误
if not token then
    error("未获取到 token，请通过环境变量传递 JWT_TOKEN")
end

-- 随机生成指定长度的字符串
function random_string(length)
    local res = ""
    for _ = 1, length do
        res = res .. string.char(math.random(97, 122))  -- 随机生成 a-z 的字符
    end
    return res
end

-- 请求构造函数，用于发布文章接口
function request()
    local publish_url = "/v1/article/publish"
    local title = random_string(10)
    local content = random_string(10000)

    local publish_body = string.format('{"title": "%s", "content": "%s"}', title, content)
    local headers = {
        ["Content-Type"] = "application/json",
        ["Authorization"] = "Bearer " .. token
    }

    return wrk.format("POST", publish_url, headers, publish_body)
end

-- 设置请求速率
function done(summary, latency, requests)
    -- 打印执行情况
    print("Completed requests: ", summary.requests)
    print("Failed requests: ", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.status)
end
