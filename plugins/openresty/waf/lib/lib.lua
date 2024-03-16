local redis_util = require "redis_util"
local action = require "action"
local cc = require "cc"
local file_utils = require "file"
local ck = require "resty.cookie"
local geo = require "geoip"
local libinjection = require "resty.libinjection"
local config = require "config"
local utils = require "utils"

local pairs = pairs
local ipairs = ipairs
local tostring = tostring
local type = type
local next = next
local concat_table = table.concat
local ngx_re_find = ngx.re.find
local ngx_re_gmatch = ngx.re.gmatch
local ngx_re_match = ngx.re.match
local ipv4_to_int = utils.ipv4_to_int
local is_ip_in_array = utils.is_ip_in_array
local is_ipv6 = utils.is_ipv6

local exec_action = action.exec_action

local _M = {}

local function is_global_state_on(name)
    return config.is_global_state_on(name)
end

local function is_site_state_on(name)
    local site_config = config.get_site_config(ngx.ctx.website_key)
    if site_config ~= nil then
        return site_config[name]["state"] == "on" and true or false
    end
    return true
end

local function is_state_on(name)
    return is_site_state_on(name) and is_global_state_on(name)
end

local function get_site_config(name)
    local site_config = config.get_site_config(ngx.ctx.website_key)
    if site_config ~= nil then
        return site_config[name]
    end
    return config.get_global_config(name)
end

local function get_site_rule(name)
    local site_rules = config.get_site_rules(ngx.ctx.website_key)
    if site_rules ~= nil then
        return site_rules[name]
    end
    return config.get_global_rules(name)
end

local function get_global_rules(name)
    return config.get_global_rules(name)
end

local function get_global_config(name)
    return config.get_global_config(name)
end

local function is_rule_state_on(rule_table)
    return rule_table["state"] == "on" and true or false
end

local function matches(input, regex, ctx, nth)
    if not nth then
        nth = 0
    end
    return ngx_re_find(input, regex, "isjo", ctx, nth)
end

local function match_rule(rule_table, str)
    if str == nil or next(rule_table) == nil then
        return false
    end
    for _, t in ipairs(rule_table) do
        if matches(str, t.rule) then
            return true, t
        end
    end

    return false
end

local function match_ip(ip_rule, ip, ipn)
    if ip_rule == nil or ip == nil then
        return false
    end
    if is_rule_state_on(ip_rule) == false then
        return false
    end
    local ip_rule_type = ip_rule.type
    if is_ipv6(ip) and ip_rule_type == "ipv6" then
        if ip == ip_rule.ipv6 then
            return true
        end
        return false
    end

    if ip_rule.type == "ipv4" then
        if ipn == ipv4_to_int(ip_rule.ipv4) then
            return true
        end
    elseif ip_rule.type == "ipArr" then
        local ip_start_n = ipv4_to_int(ip_rule.ipStart)
        local ip_end_n = ipv4_to_int(ip_rule.ipEnd)
        if is_ip_in_array(ipn, ip_start_n, ip_end_n) then
            return true
        end
    elseif ip_rule.type == "ipGroup" then
        --TODO 匹配 IP 组
    end

    return false
end

local function get_boundary()
    local header = utils.get_headers()["content-type"]
    if not header then
        return nil
    end

    if type(header) == "table" then
        header = header[1]
    end

    local m = ngx_re_match(header, ";%s*boundary=\"([^\"]+)\"")
    if m then
        return m
    end

    return ngx_re_match(header, ";%s*boundary=([^\",;]+)")
end


local function xss_and_sql_check(kv)
    if type(kv) ~= 'string' then
        return
        
    end
    if  is_site_state_on("xss") then
        local is_xss, fingerprint = libinjection.xss(tostring(kv))
        local xss_config = get_site_config("xss")
        if is_xss then
            exec_action(xss_config, { rule = kv })
            return
        end
    end
    if  is_site_state_on("sql") then
        local is_sqli, fingerprint = libinjection.sqli(tostring(kv))
        local sql_config = get_site_config("sql")
        if is_sqli then
            exec_action(sql_config, { rule = kv })
            return
        end
    end
    
end


local function get_request_body()
    ngx.req.read_body()
    local body_data = ngx.req.get_body_data()
    if not body_data then
        local body_file = ngx.req.get_body_file()
        if body_file then
            body_data = file_utils.read_file2string(body_file, true)
        end
    end
    return body_data
end

function _M.is_white_ip()
    if is_global_state_on("ipWhite") then
        local ip = ngx.ctx.ip
        if ip == "unknown" then
            return false
        end
        if ip == "127.0.0.1" then
            return true
        end
        local ipn = utils.ipv4_to_int(ip)
        local ip_rules = get_global_rules("ipWhite")
        for _, ip_rule in pairs(ip_rules) do
            if match_ip(ip_rule, ip, ipn) then
                return true
            end
        end
    end
    return false
end

function _M.allow_location_check()
    if is_state_on("geoRestrict") then
        local ip_location = ngx.ctx.ip_location
        if ip_location and ip_location.iso and ip_location.iso ~= "" then
            local iso = ip_location.iso
            local geo_config = get_site_config("geoRestrict")
            local exist = false
            for _, rule in ipairs(geo_config.rules) do
                if iso == rule then
                    exist = true
                    break

                end
            end
            local default_geo_config = {
                action = "deny",
                code = 444,
                type = "geoRestrict",
                state = "on",
                rule = iso
            }
            if exist then
                if geo_config.action == "allow" then
                    return true
                end
                if geo_config.action == "deny" then
                    exec_action(default_geo_config, default_geo_config)
                    return false
                end
            else
                if geo_config.action == "allow" then
                    exec_action(default_geo_config, default_geo_config)
                    return false
                end
            end
        end
    end
end

function _M.default_ip_black()
    if is_state_on("defaultIpBlack") then
        if geo.is_default_black_ip(ngx.ctx.ip) then
            exec_action(get_site_config("defaultIpBlack"), { rule = ngx.ctx.ip })
        end
    end
end

function _M.black_ip()
    if is_global_state_on("ipBlack") then
        local ip = ngx.ctx.ip
        if ip == "unknown" then
            return false
        end
        local exists = false

        if config.is_redis_on() then
            exists = redis_util.get("black_ip:" .. ip)
        else
            exists = ngx.shared.waf_black_ip:get(ip)
        end

        local ip_black_list = get_global_rules("ipBlack")
        local ipn = ipv4_to_int(ip)
        for _, ip_rule in pairs(ip_black_list) do
            if match_ip(ip_rule, ip, ipn) then
                exists = true
                break
            end
        end

        if exists then
            exec_action(get_global_config("ipBlack"))
        end

        return exists
    end

    return false
end

function _M.method_check()
    local method = ngx.req.get_method()
    local method_white_list = get_site_rule("methodWhite")
    for _, method_rule in ipairs(method_white_list) do
        if method_rule.rule == method and method_rule.state == 'off' then
            local method_config = get_global_config("methodWhite")
            exec_action(method_config, method_rule)
            return false
        end
    end
    return true
end

function _M.bot_check()
    if is_state_on("bot") then
        local ruri = ngx.var.request_uri
        local uri = ngx.var.uri
        local bot_rule = get_site_config("bot")
        if uri == bot_rule.uri or ruri == bot_rule.uri then
            exec_action(bot_rule)
        end
    end
end

function _M.black_ua()
    if is_global_state_on("uaBlack") then
        if type(ngx.ctx.ua) ~= 'string' then
            ngx.exit(200)
        end
        local m, mr = match_rule(get_global_rules("uaBlack"), ngx.ctx.ua)
        if m then
            exec_action(get_global_config("uaBlack"), mr)
        end
    end
end

function _M.default_ua_black()
    if is_state_on("defaultUaBlack") then
        if type(ngx.ctx.ua) ~= 'string' then
            ngx.exit(200)
        end
        local m, mr = match_rule(get_global_rules('defaultUaBlack'), ngx.ctx.ua)
        if m then
            exec_action(get_global_config('defaultUaBlack'), mr)
        end
    end
end

function _M.is_white_ua()
    if is_global_state_on("uaWhite") then
        local ua = utils.get_header("user-agent")
        if not ua then
            return false
        end
        if type(ua) ~= 'string' then
            ngx.exit(200)
        end
        for _, wa in ipairs(get_global_rules("uaWhite")) do
            if ngx.ctx.ua == wa then
                return true
            end
        end
    end
    return false
end

function _M.cc()
    if is_state_on("cc") then
        if cc.check_access_token() then
            return
        end
        local ip = ngx.ctx.ip
        local cc_config = get_site_config("cc")
        local key = ip

        if config.is_redis_on() then
            key = "cc_req_count:" .. key
            local count, _ = redis_util.incr(key, cc_config.duration)
            if not count then
                redis_util.set(key, 1, cc_config.duration)
            elseif count > cc_config.threshold then
                exec_action(cc_config, { rule = cc_config.rule })
                return
            end
        else
            local limit = ngx.shared.waf_limit
            local count, _ = limit:incr(key, 1, 0, cc_config.duration)
            if not count then
                limit:set(key, 1, cc_config.duration)
            elseif count > cc_config.threshold then
                exec_action(cc_config, { rule = cc_config.rule })
                return
            end
        end
    end
end

function _M.cc_url()
    if is_state_on("ccurl") then
        local ip = ngx.ctx.ip
        local key = ip
        local urlcc_rules = get_site_rule("ccurl")
        local urlcc_config = get_site_config("ccurl")
        local uri = ngx.var.uri

        local m, mr = match_rule(urlcc_rules, uri)
        if not m or not mr then
            return
        end
        key = uri .. key
        if config.is_redis_on() then
            key = "url_cc_req_count:" .. key
            local count, _ = redis_util.incr(key, mr.duration)
            if not count then
                redis_util.set(key, 1, mr.duration)
            elseif count > mr.threshold then
                exec_action(urlcc_config, { rule = mr.rule })
                return
            end
        else
            local limit = ngx.shared.waf_limit
            local count, _ = limit:incr(key, 1, 0, mr.duration)
            if not count then
                limit:set(key, 1, urlcc_config.duration)
            elseif count > mr.threshold then
                exec_action(urlcc_config, { rule = mr.rule })
                return
            end
        end
    end
end

function _M.is_white_url()
    if is_global_state_on("urlWhite") then
        local url = ngx.var.uri
        if url == nil or url == " " then
            return false
        end
        local m, _ = match_rule(get_global_rules("urlWhite"), url)
        if m then
            return true
        end
        return false
    end

    return false
end

function _M.black_url()
    if is_global_state_on("urlBlack") then
        local url = ngx.var.uri
        if url == nil or url == "" then
            return false
        end
        local m, mr = match_rule(get_global_rules("urlBlack"), url)
        if m then
            exec_action(get_global_config("urlBlack"), mr)
            return
        end
    end
end

function _M.default_url_black()
    if is_state_on("defaultUrlBlack") then
        local url = ngx.var.uri
        if url == nil or url == "" then
            return false
        end
        local m, mr = match_rule(get_global_rules('defaultUrlBlack'), url)
        if m then
            exec_action(get_global_config('defaultUrlBlack'), mr)
            return
        end
    end
end

function _M.args_check()
    if is_state_on("args") then
        local args = ngx.req.get_uri_args()
        if args then
            local args_list = get_global_rules("args")
            for _, val in pairs(args) do
                local val_arr = val
                if type(val) == "table" then
                    val_arr = concat_table(val, ", ")
                end
                if val_arr and type(val_arr) ~= "boolean" and val_arr ~= "" then
                    local m, mr = match_rule(args_list, utils.unescape_uri(val_arr))
                    if m then
                        exec_action(get_global_config("args"), mr)
                        return
                    end
                    xss_and_sql_check(val_arr)
                end
            end
        end
    end
end

function _M.cookie_check()
    local cookie = ngx.var.http_cookie
    if cookie and is_state_on("cookie") then
        local cookieList = get_site_rule('cookie')
        local m, mr = match_rule(cookieList, cookie)
        if m then
            exec_action(get_global_config('cookie'), mr)
            return true
        end
    end
    return false
end

function _M.header_check()
    if is_state_on("header") then
        local headers_rule = get_site_rule("header")
        local headers_config = get_site_config("header")
        local referer = ngx.var.http_referer
        if referer and referer ~= "" then
            local m = match_rule(headers_rule, referer)
            if m then
                exec_action(headers_config)
            end
        end
        local headers = utils.get_headers()
        if headers then
            for k, v in pairs(headers) do
                local m1, mr1 = match_rule(headers_rule, k)
                if m1 then
                    exec_action(headers_config, mr1)
                end
                local m2, mr2 = match_rule(headers_rule, v)
                if m2 then
                    exec_action(headers_config, mr2)
                end
            end
        end
    end
end

function _M.post_check()
    local content_type = ngx.ctx.content_type
    local content_length = ngx.ctx.content_length

    if ngx.ctx.method == "GET" or not content_type or type(content_type) ~= 'string' then
        return
    end

    if content_length == nil or content_length == 0 then
        return
    end

    local boundary =  get_boundary()

    if boundary and  is_state_on('fileExt') then
        if not ngx_re_match(content_type, '^multipart/form-data; boundary=') or  not   ngx_re_find(content_type, [[multipart]], 'ijo')then
            return
        end
        local boundary_value = ngx_re_match(content_type, '^multipart/form-data; boundary=(.+)')
        if boundary_value == nil then
            return
        end
        local data = get_request_body()
        if data == nil then
            return
        end
        local iterator = ngx_re_gmatch(data, [[Content-Disposition.+filename=.+]], 'ijo')
        if not iterator then
            return
        end

        local rule = get_site_rule("fileExt")
        while true do
            local m = iterator()
            if m then
                local match = ngx_re_match(m[0], 'Content-Disposition: form-data; (.+)filename="(.+)\\.(.*)"', 'ijo')
                if match then
                    local extension = match[3]
                    for _, ext in ipairs(rule.rules) do
                        if extension == ext then
                            exec_action(rule)
                        end
                    end
                end
            else
                break
            end
        end
    else
        ngx.req.read_body()
        local body_obj = ngx.req.get_post_args()
        if not body_obj then
            return
        end
      
        for key, val in pairs(body_obj) do
            if is_global_state_on("xss") or is_global_state_on("sql") then
                xss_and_sql_check(key)
                xss_and_sql_check(val)
            end
            if is_state_on("args") then
                local post_rules = get_global_rules("args")
                local m, mr = match_rule(post_rules, val)
                if m then
                    exec_action(get_global_config("args"), mr)
                    return
                end
            end
        end
    end

end

function _M.acl()
    local rules = get_site_rule("acl")
    for _, rule in pairs(rules) do
        if rule.state == nil or rule.state == "off" then
            goto continue
        end
        local conditions = rule.conditions
        local match = true
        for _, condition in pairs(conditions) do
            local field = condition.field
            local field_name = condition.name
            local pattern = condition.pattern
            local match_value = ''
            if field == 'URL' then
                match_value = ngx.var.request_uri

            elseif field == 'Cookie' then
                if field_name ~= nil and field_name ~= '' then
                    local cookies, _ = ck:new()
                    if not cookies then
                        match = false
                        break
                    else
                        match_value, _ = cookies:get(field_name)
                    end
                else
                    match_value = ngx.var.http_cookie
                end

            elseif field == 'Header' then
                local headers = ngx.req.get_headers()
                if headers then
                    if field_name ~= nil and field_name ~= '' then
                        match_value = headers[field_name]
                    else
                        match_value = concat_table(headers, '')
                    end
                else
                    match = false
                    break
                end

            elseif field == 'Referer' then
                match_value = ngx.var.http_referer

            elseif field == 'User-Agent' then
                match_value = ngx.var.http_user_agent

            elseif field == 'IP' then
                match_value = ngx.ctx.ip
            end

            if pattern == '' then
                if match_value ~= nil and match_value ~= '' then
                    match = false
                    break
                end
            else
                if not matches(match_value, pattern) then
                    match = false
                    break
                end
            end
        end
        if match then
            rule.type = "acl"
            exec_action(rule)
        end
        :: continue ::
    end
end

return _M
