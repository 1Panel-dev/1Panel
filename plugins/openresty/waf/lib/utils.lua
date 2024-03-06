local geoip = require "geoip"
local sub_str = string.sub
local pairs = pairs
local insert_table = table.insert
local tonumber = tonumber
local ipairs = ipairs
local type = type
local find_str = string.find
local gmatch_str = string.gmatch
local cjson = require "cjson"

local _M = {}

function _M.split(input_string, delimiter)
    local result = {}
    for part in input_string:gmatch("([^" .. delimiter .. "]+)") do
        insert_table(result, part)
    end
    return result
end

function _M.get_cookie_list(cookie_str)
    local cookies = {}
    for cookie in cookie_str:gmatch("([^;]+)") do
        local key, value = cookie:match("^%s*([^=]+)=(.*)$")
        if key and value then
            cookies[key] = value
        end
    end
    return cookies
end

function _M.unescape_uri(str)
    local newStr = str
    for t = 1, 2 do
        local temp = ngx.unescape_uri(newStr)
        if not temp then
            break
        end
        newStr = temp
    end
    return newStr
end

function _M.get_expire_time()
    local localtime = ngx.localtime()
    local hour = sub_str(localtime, 12, 13)
    local expire_time = (24 - tonumber(hour)) * 3600
    return expire_time
end

function _M.get_date_hour()
    local localtime = ngx.localtime()
    local hour = sub_str(localtime, 1, 13)
    return hour
end

function _M.getHours()
    local hours = {}
    local today = ngx.today()
    local hour = nil
    for i = 0, 23 do
        if i < 10 then
            hour = today .. ' 0' .. i
        else
            hour = today .. ' ' .. i
        end
        hours[i + 1] = hour
    end

    return hours
end

function _M.ipv4_to_int(ip)
    local ipInt = 0
    for i, octet in ipairs({ ip:match("(%d+)%.(%d+)%.(%d+)%.(%d+)") }) do
        ipInt = ipInt + tonumber(octet) * 256 ^ (4 - i)
    end
    return ipInt
end

function _M.is_ipv6(ip)
    if find_str(ip, ':') then
        return true
    end
    return false
end

function _M.is_ip_in_array(ip, ipStart, ipEnd)
    if ip >= ipStart and ip <= ipEnd then
        return true
    end
    return false
end

function _M.get_real_ip()
    local var = ngx.var
    local ips = {
        var.http_x_forwarded_for,
        var.http_proxy_client_ip,
        var.http_wl_proxy_client_ip,
        var.http_http_client_ip,
        var.http_http_x_forwarded_for,
        var.remote_addr
    }

    for _, ip in pairs(ips) do
        if ip and ip ~= "" then
            if type(ip) == "table" then
                ip = ip[1]
            end
            return ip
        end
    end

    return "unknown"
end

function _M.get_geo_ip(ip)
    if _M.is_intranet_address(ip) then
        return {
            country = { ["zh"] = "内网", ["en"] = "intranet" },
            province = { ["zh"] = "内网", ["en"] = "intranet" },
            city = { ["zh"] = "内网", ["en"] = "intranet" },
            longitude = 0,
            latitude = 0,
            iso = "local"
        }
    else
        geoip.init()
        local geo_res = geoip.lookup(ip)
        local msg = "访问 IP  " .. ip
        if geo_res.country then
            msg = msg .. " 国家 " .. cjson.encode(geo_res.country)
        end
        if geo_res.province then
            msg = msg .. " 省份 " .. cjson.encode(geo_res.province)
        end
        ngx.log(ngx.ERR, msg)
        return  geo_res
        
    end
end

function _M.get_header(headerKey)
    return ngx.req.get_headers(20000)[headerKey]
end

function _M.get_headers()
    return ngx.req.get_headers(20000)
end

function _M.is_intranet_address(ip_addr)
    if not ip_addr then
        return false
    end
    if ip_addr == "unknown" then
        return false
    end
    if find_str(ip_addr, ':') then
        return false
    end
    
    local parts = {}
    for part in gmatch_str(ip_addr, "%d+") do
        insert_table(parts, tonumber(part))
    end
    if parts[1] == 10 or
            (parts[1] == 192 and parts[2] == 168) or
            (parts[1] == 172 and parts[2] >= 16 and parts[2] <= 31) then
        return true
    else
        return false
    end
end

function _M.get_wafdb(waf_db_path)
    local ok, sqlite3 = pcall(function()
        return require "lsqlite3"
    end)
    if not ok then
        return nil
    end
    return sqlite3.open(waf_db_path)
end
return _M
