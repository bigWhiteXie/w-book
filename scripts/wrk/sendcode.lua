-- login_sms.lua
local headers = {
    ["Content-Type"] = "application/json"
}

-- 设置随机种子，确保每次运行随机数不同
math.randomseed(os.time())

-- 生成一个随机手机号码（9 位数字）
function generate_phone()
    local phone_number = tostring(math.random(100000000, 999999999))
    return phone_number
end

function request()
    -- 随机生成一个手机号码
    local phone = generate_phone()

    -- 构建请求体
    local body = string.format('{"phone": "%s"}', phone)

    -- 发送 POST 请求
    return wrk.format("POST", "/v1/user/login_sms/code", headers, body)
end
