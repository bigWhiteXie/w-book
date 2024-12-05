-- publish_resources.lua
local ids = {}


-- 随机生成标题和内容
local function random_string(length)
    local charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    local str = ""
    for i = 1, length do
        local index = math.random(1, #charset)
        str = str .. charset:sub(index, index)
    end
    return str
end

function request()
    local body = string.format('{"title": "%s", "content": "%s"}', random_string(10), random_string(300))
    local headers = {
        ["Content-Type"] = "application/json",
        ["Authorization"] = "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjk4Njc1MzMsIlVpZCI6MSwiU3NpZCI6IjlmOTVlM2Q5LTkyMTYtMTFlZi1hNmRhLTAwMGMyOWU1Yjk3OSIsIlVzZXJBZ3JlbnQiOiJBcGlmb3gvMS4wLjAgKGh0dHBzOi8vYXBpZm94LmNvbSkifQ.FTxXS8EYMmdcJ8Hhza_4XWiPEN2x4exG1G3IizGSyw_Oho8WQFBZbdEYxdHYKOXSyKmss9l9_QcHk5IATatc4A"
    }
    return wrk.format("POST", "/v1/article/publish", headers, body)
end

function response(status, headers, body)
    if status == 200 then
        local json = require("cjson").decode(body)
        table.insert(ids, json.data)
    end
end

-- 在测试完成后保存 ID 列表到文件
function done(summary, latency, requests)
    local file = io.open("resource_ids.txt", "w")
    for _, id in ipairs(ids) do
        file:write(id .. "\n")
    end
    file:close()
end

function decode_json(s)
    local json = {}
    for k, v in s:gmatch('"([^"]+)":%s*"([^"]+)"') do
        json[k] = v
    end
    return json
end
