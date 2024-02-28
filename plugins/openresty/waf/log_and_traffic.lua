local utils = require "utils"
local stringutf8 = require "stringutf8"
local logger_factory = require "logger_factory"
local db = require "db"
local config = require "config"
local redis_util = require "redis_util"
local action = require "action"

local upper_str = string.upper
local concat_table = table.concat
local tonumber = tonumber
local get_expire_time = utils.get_expire_time
local get_date_hour = utils.get_date_hour
local get_today = ngx.today

local ATTACK_PREFIX = "attack_"
local ATTACK_TYPE_PREFIX = "attack_type_"

local function writeAttackLog()
    local rule_table = ngx.ctx.rule_table
    local data = ngx.ctx.hitData
    local action = ngx.ctx.action
    local rule = rule_table.rule

    local rule_type = rule_table.type
    if not rule_type then
        rule_type = "default"
    end

    local realIp = ngx.ctx.ip

    local geoip = ngx.ctx.geoip
    local country = geoip.country["zh"] or ""
    local province = geoip.province["zh"] or ""
    local city = ""
    local longitude = geoip.longitude
    local latitude = geoip.latitude

    local method = ngx.req.get_method()
    local uri = ngx.var.request_uri
    local ua = ngx.ctx.ua
    local host = ngx.var.server_name
    local protocol = ngx.var.server_protocol
    local attackTime = ngx.localtime()

    local website_key = ngx.ctx.website_key

    local address = country .. province .. city
    address = stringutf8.default_if_blank(address, '-')
    ua = stringutf8.default_if_blank(ua, '-')
    data = stringutf8.default_if_blank(data, '-')

    local log_path = "/www/sites/" .. website_key .. "/attack.log"
    local logStr = concat_table({ rule_type, realIp, address, "[" .. attackTime .. "]", '"' .. method, host, uri, protocol .. '"', data, '"' .. ua .. '"', '"' .. rule .. '"', action }, ' ')
    local host_logger = logger_factory.get_logger(log_path, host, true)
    host_logger:log(logStr .. '\n')

    db.init_db()
    if wafdb == nil then
        return false
    end

    local isBlock = 0
    local blocking_time = 0
    if ngx.ctx.ipBlocked then
        isBlock = 1
        blocking_time = tonumber(rule_table.ipBlockTime)
    end

    local insertQuery = [[
        INSERT INTO attack_log (
            ip, ip_city, ip_country, ip_subdivisions, ip_continent,
            ip_longitude, ip_latitude, time, localtime, server_name,
            website_key, host, method, uri, user_agent, rule,
            nginx_log, blocking_time, action, msg, params, is_block
        ) VALUES (
            :realIp, :city, :country, :subdivisions, :continent,
            :longitude, :latitude, :time, :localtime, :host,
            :website_key, :host, :method, :uri, :ua, :rule_type,
            :logStr, :blocking_time, :action, :msg, :params, :is_block
        )
     ]]

    local stmt = wafdb:prepare(insertQuery)

    stmt:bind_names {
        realIp = realIp,
        city = city,
        country = country,
        subdivisions = "",
        continent = "",
        longitude = longitude,
        latitude = latitude,
        time = os.time(),
        localtime = os.date("%Y-%m-%d %H:%M:%S", os.time()),
        host = host,
        website_key = website_key,
        method = method,
        uri = uri,
        ua = ua,
        rule_type = rule_type,
        logStr = logStr,
        blocking_time = blocking_time or 0,
        action = action,
        msg = "msg",
        params = "Params",
        is_block = isBlock
    }

    local code = stmt:step()
    stmt:finalize()

    if code ~= 101 then
        local errorMsg = wafdb:errmsg()
        if errorMsg then
            ngx.log(ngx.ERR, "insert attack_log error", errorMsg)
        end
    end

end

local function writeIPBlockLog()
    local rule_table = ngx.ctx.rule_table
    local ip = ngx.ctx.ip
    local website_key = ngx.ctx.website_key
    local log_path = "/www/sites/" .. website_key .. "/attack.log"
    local host_logger = logger_factory.get_logger(log_path .. "ipBlock.log", 'ipBlock', false)

    host_logger:log(concat_table({ ngx.localtime(), ip, rule_table.type, rule_table.ipBlockTime .. 's' }, ' ') .. "\n")

    --todo 永久拉黑IP
    --if rule_table.ipBlockTimeout == 0 then
    --    local ipBlackLogger = logger_factory.get_logger(rulePath .. "ipBlackList", 'ipBlack', false)
    --    ipBlackLogger:log(ip .. "\n")
    --end
end

-- 按小时统计当天请求流量，存入缓存，key格式：2023-05-05 09
local function countRequestTraffic()
    local hour = get_date_hour()
    local dict = ngx.shared.dict_req_count
    local expire_time = get_expire_time()
    local count, err = dict:incr(hour, 1, 0, expire_time)
    if not count then
        dict:set(hour, 1, expire_time)
        ngx.log(ngx.ERR, "failed to count traffic ", err)
    end
end

--[[
    按小时统计当天攻击请求流量，存入缓存，key格式：attack_2023-05-05 09
    按天统计当天所有攻击类型流量，存入缓存，key格式：attack_type_2023-05-05_ARGS
]]
local function countAttackRequestTraffic()
    local rule_table = ngx.ctx.rule_table
    local rule_type = ""
    if rule_table.rule_type then
        rule_type = upper_str(rule_table.rule_type)
    end
    if rule_table.type then
        rule_type = upper_str(rule_table.type)
    end
    local dict = ngx.shared.dict_req_count
    local count, err = nil, nil
    local expire_time = get_expire_time()

    if rule_type ~= 'WHITEIP' then
        local hour = get_date_hour()
        local key = ATTACK_PREFIX .. hour
        count, err = dict:incr(key, 1, 0, expire_time)
        if not count then
            dict:set(key, 1, expire_time)
            ngx.log(ngx.ERR, "failed to count attack traffic ", err)
        end
    end

    local today = get_today() .. '_'
    local type_key = ATTACK_TYPE_PREFIX .. today .. rule_type
    count, err = dict:incr(type_key, 1, 0, expire_time)

    if not count and err == "not found" then
        dict:set(type_key, 1, expire_time)
        ngx.log(ngx.ERR, "failed to count attack traffic ", err)
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
    countRequestTraffic()
    local isAttack = ngx.ctx.isAttack

    if isAttack then
        writeAttackLog()
        countAttackRequestTraffic()
    end

    -- if ngx.ctx.ipBlocked then
    --     writeIPBlockLog()
    -- end
end
