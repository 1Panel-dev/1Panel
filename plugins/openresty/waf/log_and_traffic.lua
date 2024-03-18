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
    local wafdb = utils.get_wafdb(config.waf_log_db_path)
    if not wafdb then
        ngx.log(ngx.ERR, "get log db failed")
        return 
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

    local exec_rule = {}
    local rule_action = ""
    local exec_rule_type = ""
    local match_rule_detail = ""
    local match_rule_type = ""
    local is_attack = 0
    local is_block = 0
    local blocking_time = 0
    
    local method = ""
    local uri = ""
    local ua = ""
    local host = ""
    local protocol = ""
    local website_key = ""
    local logs_str = ""

    if attack then
        exec_rule = ngx.ctx.exec_rule
        rule_action = exec_rule.action
        exec_rule_type = exec_rule.type
        is_attack = 1
        method = ngx.req.get_method()
        uri = ngx.var.request_uri
        ua = ngx.ctx.ua
        host = ngx.var.server_name
        protocol = ngx.var.server_protocol or ""
        website_key = ngx.ctx.website_key
        
        if exec_rule.match_rule then
            match_rule_detail = exec_rule.match_rule.rule
            match_rule_type = exec_rule.match_rule.type
        end
     
        if ngx.ctx.ip_blocked then
            is_block = 1
            blocking_time = tonumber(exec_rule.ipBlockTime)
        end

        logs_str = method .. "  " .. uri .. " "..protocol.."\n"
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
    end

    local log_id = uuid()
    local insertQuery = [[
        INSERT INTO req_logs (
            id, ip, ip_iso, ip_country_zh, ip_country_en,
            ip_province_zh, ip_province_en, ip_longitude, ip_latitude,
            localtime, server_name,  website_key, host, method, 
            uri, user_agent, exec_rule, rule_type, match_rule, match_value,
            nginx_log, blocking_time, action, is_block,is_attack
        ) VALUES (
            :id, :real_ip, :iso,  :country_zh, :country_en,
            :province_zh, :province_en,:longitude, :latitude,
            DATETIME('now'), :server_name,:host, :website_key, :method, 
            :uri, :ua,  :exec_rule, :rule_type, :match_rule, :match_value,
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
        exec_rule = exec_rule_type,
        rule_type = match_rule_type,
        match_rule = match_rule_detail,
        match_value = "",
        logs_str = logs_str, 
        blocking_time = blocking_time or 0,
        action = rule_action,
        is_block = is_block,
        is_attack = is_attack
    }
    stmt:step()
    stmt:finalize()

    local code2 = 101
    if ngx.ctx.ip_blocked then
        local insertBlockIp = [[
            INSERT INTO block_ips (ip, is_block, blocking_time, req_log_id,create_date)
            VALUES (:ip, :is_block, :blocking_time, :req_log_id, DATETIME('now'))
        ]]
        stmt = wafdb:prepare(insertBlockIp)
        stmt:bind_names {
            ip=real_ip,
            is_block = is_block,
            blocking_time = blocking_time or 0,
            req_log_id = log_id
        }
        code2 = stmt:step()
        stmt:finalize()
    end

    wafdb:execute([[COMMIT]])
    
    --local error_msg = wafdb:errmsg()
    --if error_msg then
    --    ngx.log(ngx.ERR, "insert attack_log error ", error_msg .. "  ")
    --end

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

local add_count = function(shared_dict,key)
    local count, _ = shared_dict:incr(key, 1)
    if not count then
        shared_dict:set(key, 1)
    end
end

local function count_req_status(is_attack)
    local status = ngx.status
    local req_count = ngx.shared.dict_req_count
    add_count(req_count, "req_count")
    if (status >= 400 and status < 500) then
        add_count(req_count, "count_4xx")
    end
    if (status >= 500) then
        add_count(req_count, "count_5xx")
    end
    if is_attack then
        add_count(req_count, "attack_count")
    end
end

if config.is_waf_on() then
    if ngx.ctx.is_waf_url then
        return
    end
    count_not_found()
    local is_attack = ngx.ctx.is_attack
    
    if not ngx.ctx.ip then
        ngx.ctx.ip = utils.get_real_ip()
        ngx.ctx.ip_location = utils.get_ip_location(ngx.ctx.ip)
    end
    
    count_req_status(is_attack)
    if is_attack then
        write_req_log(is_attack)
    end
end
