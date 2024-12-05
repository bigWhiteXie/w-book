-- test_top.lua
local headers = {
    ["Authorization"] = "Bearer YOUR_TOKEN_HERE" -- 替换为你的实际 token
}

function request()
    return wrk.format("GET", "/v1/top/article", headers)
end
