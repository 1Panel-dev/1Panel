local utils = require "utils"
local config = require "config"
local redis_util = require "redis_util"
local action = require "action"
local uuid = require"resty.uuid"

local upper_str = string.upper
local tonumber = tonumber
local pairs = pairs


local function writeAttackLog()
    local rule_table = ngx.ctx.rule_table
    local action = ngx.ctx.action
    local rule = rule_table.rule

    local rule_type = rule_table.type
    if not rule_type then
        rule_type = "default"
    end

    local real_ip = ngx.ctx.ip
    local geoip = ngx.ctx.geoip
    local country = geoip.country
    if not country then
        country["zh"] = "unknown"
        country["en"] = "unknown"
    end
    local province = geoip.province
    if not province then
        province["zh"] = "unknown"
        province["en"] = "unknown"
    end
    local longitude = geoip.longitude
    local latitude = geoip.latitude
    local iso = geoip.iso

    local method = ngx.req.get_method()
    local uri = ngx.var.request_uri
    local ua = ngx.ctx.ua
    local host = ngx.var.server_name
    local protocol = ngx.var.server_protocol
    local website_key = ngx.ctx.website_key
 

   
    local logs_str = method .. "  " .. uri .. " "..protocol.."\n"
    local headers = ngx.req.get_headers(20000)
    for k, v in pairs(headers) do
        local value = ""
        if v then
            value = v
        end
        logs_str = logs_str .. upper_str(k) .. ": " .. value .. "\n"
    end
    

    local isBlock = 0
    local blocking_time = 0
    if ngx.ctx.ipBlocked then
        isBlock = 1
        blocking_time = tonumber(rule_table.ipBlockTime)
    end

    local log_id = uuid()
    local time = os.time()
    local localtime = os.date("%Y-%m-%d %H:%M:%S", time)


    local wafdb = utils.get_wafdb(config.waf_db_path)
    if wafdb == nil then
        return false
    end
    
    local insertQuery = [[
        INSERT INTO attack_log (
            id, ip, ip_iso, ip_country_zh, ip_country_en,
            ip_province_zh, ip_province_en, ip_longitude, ip_latitude,
            time, localtime, server_name,  website_key, host, method, 
            uri, user_agent, rule_type,match_rule, match_value,
            nginx_log, blocking_time, action, is_block
        ) VALUES (
            :id, :real_ip, :iso,  :country_zh, :country_en,
            :province_zh, :province_en,:longitude, :latitude,
            :time, :localtime, :server_name,:host, :website_key, :method, 
            :uri, :ua, :rule_type, :match_rule, :match_value,
            :logs_str, :blocking_time, :action, :is_block
        )
     ]]

    local stmt = wafdb:prepare(insertQuery)

    stmt:bind_names {
        id = log_id,
        iso = iso,
        real_ip = real_ip,
        country_zh = country["zh"],
        country_en = country["en"],
        province_zh = province["zh"],
        province_en = province["en"],
        longitude = longitude,
        latitude = latitude,
        time = time,
        localtime = localtime,
        host = host,
        server_name = host,
        website_key = website_key,
        method = method,
        uri = uri,
        ua = ua,
        rule_type = rule_type,
        match_rule = rule,
        match_value = "",
        logs_str = logs_str, 
        blocking_time = blocking_time or 0,
        action = action,
        is_block = isBlock
    }

    local code = stmt:step()
    stmt:finalize()

    if code ~= 101 then
        local errorMsg = wafdb:errmsg()
        if errorMsg then
            ngx.log(ngx.ERR, "insert attack_log error ", errorMsg .. "  ")
        end
    end

end

local function count_not_found()
    if ngx.status ~= 404 then
        return
    end
    if config.is_global_state_on("notFoundCount") then
        local ip = ngx.ctx.ip
        local not_found_config = config.get_global_config("notFoundCount")
        local key = ip

        if config.is_redis_on() then
            key = "cc_attack_count:" .. key
            local count, _ = redis_util.incr(key, not_found_config.duration)
            if not count then
                redis_util.set(key, 1, not_found_config.duration)
            elseif count >= not_found_config.threshold then
                action.block_ip(ip, not_found_config)
                return
            end
        else
            key = ip .. "not_found"
            local limit = ngx.shared.waf_limit
            local count, _ = limit:incr(key, 1, 0, not_found_config.duration)
            if not count then
                limit:set(key, 1, not_found_config.duration)
            elseif count >= not_found_config.threshold then
                action.block_ip(ip, not_found_config)
                return
            end
        end
    end
end

if config.is_waf_on() then
    count_not_found()
    local isAttack = ngx.ctx.isAttack

    if isAttack then
        writeAttackLog()
    end
end
