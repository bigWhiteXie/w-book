-- like_resources.lua

local headers = {
    ["Content-Type"] = "application/json",
    ["Authorization"] = "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjk4Njc1MzMsIlVpZCI6MSwiU3NpZCI6IjlmOTVlM2Q5LTkyMTYtMTFlZi1hNmRhLTAwMGMyOWU1Yjk3OSIsIlVzZXJBZ3JlbnQiOiJBcGlmb3gvMS4wLjAgKGh0dHBzOi8vYXBpZm94LmNvbSkifQ.FTxXS8EYMmdcJ8Hhza_4XWiPEN2x4exG1G3IizGSyw_Oho8WQFBZbdEYxdHYKOXSyKmss9l9_QcHk5IATatc4A" -- 替换为你的实际 token
}

-- 设置随机种子，确保每次运行随机数不同
math.randomseed(os.time())

function request()
    -- 随机生成一个 1 到 50000 的数字作为资源 ID
    local id = math.random(1, 50000)

    -- 根据奇偶性决定点赞数
    local like_count = (id % 2 == 0) and 2 or 1

    -- 构造请求体
    local body = string.format('{"biz": "article", "biz_id": %d, "action": %d}', id, 1)

    -- 返回 POST 请求
    return wrk.format("POST", "/v1/resource/like", headers, body)
end
