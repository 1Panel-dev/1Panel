local utils = require "utils"
local config = require "config"
local redis_util = require "redis_util"
local action = require "action"
local uuid = require"resty.uuid"

local upper_str = string.upper
local tonumber = tonumber
local pairs = pairs
local type = type
local concat_table = table.concat

local function write_req_log(attack)
    local wafdb = utils.get_wafdb(config.waf_log_path)
    if not wafdb then
        ngx.log(ngx.ERR, "get log db failed")
        return 
    end
    
    local rule_table = nil
    local rule_action = ""
    local rule = nil
    local rule_type = ""
    local is_attack = 0
    local is_block = 0
    local blocking_time = 0
 
    if attack then
        rule_table = ngx.ctx.rule_table
        rule_action = ngx.ctx.action
        rule = rule_table.rule
        rule_type = rule_table.type
        is_attack = 1
        if not rule_type then
            rule_type = "default"
        end
        if ngx.ctx.ip_blocked then
            is_block = 1
            blocking_time = tonumber(rule_table.ipBlockTime)
        end
    end

    local real_ip = ngx.ctx.ip
    local ip_location = ngx.ctx.ip_location
    local country
    local province
    local longitude = 0.0
    local latitude = 0.0
    local iso = "CN"
    if ip_location then
        country = ip_location.country or {
            ["zh"] = "unknown",
            ["en"] = "unknown"
        }
        province = ip_location.province or {
            ["zh"] = "",
            ["en"] = ""
        }
        longitude = ip_location.longitude
        latitude = ip_location.latitude
        iso = ip_location.iso
    end
    
    local method = ngx.req.get_method()
    local uri = ngx.var.request_uri
    local ua = ngx.ctx.ua
    local host = ngx.var.server_name
    local protocol = ngx.var.server_protocol or ""
    local website_key = ngx.ctx.website_key
    
    local logs_str = method .. "  " .. uri .. " "..protocol.."\n"
    local headers = ngx.req.get_headers(20000)
    for k, v in pairs(headers) do
        local value = ""
        if v then
            if type(v) == "table" then
                value = concat_table(v, ",")
            else
                value = v
            end
        end
        logs_str = logs_str .. upper_str(k) .. ": " .. value .. "\n"
    end


    local log_id = uuid()
    local insertQuery = [[
        INSERT INTO req_logs (
            id, ip, ip_iso, ip_country_zh, ip_country_en,
            ip_province_zh, ip_province_en, ip_longitude, ip_latitude,
            localtime, server_name,  website_key, host, method, 
            uri, user_agent, rule_type,match_rule, match_value,
            nginx_log, blocking_time, action, is_block,is_attack
        ) VALUES (
            :id, :real_ip, :iso,  :country_zh, :country_en,
            :province_zh, :province_en,:longitude, :latitude,
            DATETIME('now'), :server_name,:host, :website_key, :method, 
            :uri, :ua, :rule_type, :match_rule, :match_value,
            :logs_str, :blocking_time, :action, :is_block, :is_attack
        )
     ]]

    wafdb:execute([[BEGIN TRANSACTION]])
    
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
        action = rule_action,
        is_block = is_block,
        is_attack = is_attack
    }
    local code = stmt:step()
    stmt:finalize()

    local code2 = 101
    if ngx.ctx.ip_blocked then
        local insertBlockIp = [[
            INSERT INTO block_ip (ip, is_block, blocking_time, attack_log_id)
            VALUES (:ip, :is_block, :blocking_time, :attack_log_id)
        ]]
        stmt = wafdb:prepare(insertBlockIp)
        stmt:bind_names {
            ip=real_ip,
            is_block = is_block,
            blocking_time = blocking_time or 0,
            attack_log_id = log_id
        }
        code2 = stmt:step()
        stmt:finalize()
    end

    wafdb:execute([[COMMIT]])
    
    if code ~= 101 or code2 ~= 101 then
        local error_msg = wafdb:errmsg()
        if error_msg then
            ngx.log(ngx.ERR, "insert attack_log error ", error_msg .. "  ")
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

local function count_req_status(is_attack)
    local wafdb = utils.get_wafdb(config.waf_log_db_path)
    if not wafdb then
        ngx.log(ngx.ERR, "get log db failed")
        return
    end
    
    local today = ngx.today()
    local status = ngx.status

    local stmt_exist = wafdb:prepare("SELECT COUNT(*) FROM waf_stat WHERE day = ?")
    stmt_exist:bind_values(today)
    stmt_exist:step()
    local count = stmt_exist:get_uvalues()
    stmt_exist:finalize()
    
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
        local error_msg = wafdb:errmsg()
        if error_msg then
            ngx.log(ngx.ERR, "update waf_stat error ", error_msg .. "  ")
        end
    end
end

if config.is_waf_on() then
    count_not_found()
    local is_attack = ngx.ctx.is_attack
    
    if not ngx.ctx.ip then
        ngx.ctx.ip = utils.get_real_ip()
        ngx.ctx.ip_location = utils.get_ip_location(ngx.ctx.ip)
        local ua = utils.get_header("user-agent")
        if not ua then
            ua = ""
        end
    end
    
 
    count_req_status(is_attack)
    write_req_log(is_attack)
end
