local utils = require "utils"
local config = require "config"
local redis_util = require "redis_util"
local action = require "action"
local uuid = require"resty.uuid"

local upper_str = string.upper
local tonumber = tonumber
local pairs = pairs

local function write_req_log(wafdb,attack)
    local rule_table = nil
    local action = ""
    local rule = nil
    local rule_type = ""
    local is_attack = 0
    local is_block = 0
    local blocking_time = 0
 
    if attack then
        rule_table = ngx.ctx.rule_table
        action = ngx.ctx.action
        rule = rule_table.rule
        rule_type = rule_table.type
        is_attack = 1
        if not rule_type then
            rule_type = "default"
        end
        if ngx.ctx.ipBlocked then
            is_block = 1
            blocking_time = tonumber(rule_table.ipBlockTime)
        end
    end

    local real_ip = ngx.ctx.ip
    local geoip = ngx.ctx.geoip
    local country
    local province
    local longitude = 0.0
    local latitude = 0.0
    local iso = "CN"
    if geoip then
        country = geoip.country
        province = geoip.province
        longitude = geoip.longitude
        latitude = geoip.latitude
        iso = geoip.iso
    else 
        ngx.log(ngx.ERR, real_ip .. " 无法获取地址")
    end
    if not country then
        country = {
            ["zh"] = "unknown",
            ["en"] = "unknown"
        }
    end
    if not province then
        province = {
            ["zh"] = "unknown",
            ["en"] = "unknown"
        }
    end

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
            if type(v) == "table" then
                value = table.concat(v, ",")
            else
                value = v
            end
        end
        logs_str = logs_str .. upper_str(k) .. ": " .. value .. "\n"
    end

    local log_id = uuid()
    local time = os.time()
    local localtime = os.date("%Y-%m-%d %H:%M:%S", time)
    
    local insertQuery = [[
        INSERT INTO req_logs (
            id, ip, ip_iso, ip_country_zh, ip_country_en,
            ip_province_zh, ip_province_en, ip_longitude, ip_latitude,
            time, localtime, server_name,  website_key, host, method, 
            uri, user_agent, rule_type,match_rule, match_value,
            nginx_log, blocking_time, action, is_block,is_attack
        ) VALUES (
            :id, :real_ip, :iso,  :country_zh, :country_en,
            :province_zh, :province_en,:longitude, :latitude,
            :time, :localtime, :server_name,:host, :website_key, :method, 
            :uri, :ua, :rule_type, :match_rule, :match_value,
            :logs_str, :blocking_time, :action, :is_block, :is_attack
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
        is_block = is_block,
        is_attack = is_attack
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

local function count_req_status(wafdb,is_attack)
    local today = ngx.today()
    local status = ngx.status

    local stmt_exist = wafdb:prepare("SELECT COUNT(*) FROM waf_stat WHERE day = ?")
    stmt_exist:bind_values(today)
    stmt_exist:step()
    local count = stmt_exist:get_uvalues()

    local req_count_update = 1
    local count_4xx_update = (status >= 400 and status < 500) and 1 or 0
    local count_5xx_update = (status >= 500) and 1 or 0
    local attack_count_update = is_attack and 1 or 0
    local code = 0
    
    if count > 0 then
        local stmt = wafdb:prepare("UPDATE waf_stat SET req_count = req_count + ?, count4xx = count4xx + ?, count5xx = count5xx + ?, attack_count = attack_count + ? WHERE day = ?")
        stmt:bind_values(req_count_update, count_4xx_update, count_5xx_update, attack_count_update, today)
        code = stmt:step()
        stmt:finalize()
    else
        local stmt = wafdb:prepare("INSERT INTO waf_stat (day, req_count, count4xx, count5xx, attack_count,create_date) VALUES (?, ?, ?, ?, ?,DATETIME('now'))")
        stmt:bind_values(today, req_count_update, count_4xx_update, count_5xx_update, attack_count_update)
        code = stmt:step()
        stmt:finalize()
    end

    if code ~= 101 then
        local errorMsg = wafdb:errmsg()
        if errorMsg then
            ngx.log(ngx.ERR, "update waf_stat error ", errorMsg .. "  ")
        end
    end
end

if config.is_waf_on() then
    count_not_found()
    local is_attack = ngx.ctx.is_attack
    
    if not ngx.ctx.ip then
        ngx.ctx.ip = utils.get_real_ip()
        ngx.ctx.geoip = utils.get_geo_ip(ngx.ctx.ip)
        local ua = utils.get_header("user-agent")
        if not ua then
            ua = ""
        end
    end
    
    
    local wafdb = utils.get_wafdb(config.waf_db_path)
    if wafdb ~= nil then
        count_req_status(wafdb,is_attack)
        write_req_log(wafdb,is_attack)
    end
end
