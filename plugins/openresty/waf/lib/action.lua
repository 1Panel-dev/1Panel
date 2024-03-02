local config = require "config"
local redis_util = require "redis_util"
local format_str = string.format

local _M = {}

local function deny(status_code, res)
    ngx.status = status_code
    if res then
        ngx.header.content_type = "text/html; charset=UTF-8"
        ngx.say(config.get_html_res(res))
    end
    ngx.exit(ngx.status)
end

local function redirect(status_code)
    ngx.header.content_type = "text/html; charset=UTF-8"
    ngx.say(config.get_html_res("redirect"))
    ngx.status = status_code
    ngx.exit(ngx.status)
end

local function slide()
    ngx.header.content_type = "text/html; charset=UTF-8"
    ngx.header.Cache_Control = "no-cache"
    ngx.status = 200
    ngx.say(format_str(config.get_html_res("slide"), ngx.md5(ngx.ctx.ip)))
    ngx.exit(ngx.status)
end

local function five_second()
    ngx.header.content_type = "text/html; charset=UTF-8"
    ngx.header.Cache_Control = "no-cache"
    ngx.status = 200
    ngx.say(format_str(config.get_html_res("five_second"), ngx.md5(ngx.ctx.ip)))
    ngx.exit(ngx.status)
end

function _M.block_ip(ip, rule)
    local ok, err = nil, nil
    local msg = "拉黑IP :  " .. ip .. "国家 " .. ngx.ctx.geoip.country["zh"]
    if rule then
        msg = msg .. " 规则 " .. rule.type
    end

    ngx.log(ngx.ERR, msg)

    if config.redis_on then
        local red, err1 = redis_util.get_conn()
        if not red then
            return nil, err1
        end
        local key = "black_ip:" .. ip

        local exists = red:exists(key)
        if exists == 0 then
            ok, err = red:set(key, 1)
            if ok then
                ngx.ctx.ipBlocked = true
            else
                ngx.log(ngx.ERR, "failed to set redis key " .. key, err)
            end
        end

        if rule.ipBlockTime > 0 then
            ok, err = red:expire(key, rule.ipBlockTime)
            if not ok then
                ngx.log(ngx.ERR, "failed to expire redis key " .. key, err)
            end
        end

        redis_util.close_conn(red)
    else
        local wafBlackIp = ngx.shared.waf_black_ip
        local exists = wafBlackIp:get(ip)
        if not exists then
            ok, err = wafBlackIp:set(ip, 1, rule.ipBlockTime)
            if ok then
                ngx.ctx.ipBlocked = true
            else
                ngx.log(ngx.ERR, "failed to set key " .. ip, err)
            end
        elseif rule.ipBlockTime > 0 then
            ok, err = wafBlackIp:expire(ip, rule.ipBlockTime)
            if ok then
                ngx.ctx.ipBlocked = true
            else
                ngx.log(ngx.ERR, "failed to expire key " .. ip, err)
            end
        end
    end

    return ok
end

local function attack_count(config_type)
    if config_type == "ipBlack" then
        return
    end
    if config.is_global_state_on("attackCount") then
        local ip = ngx.ctx.ip
        local attack_config = config.get_global_config("attackCount")
        local key = ip

        if config.is_redis_on() then
            key = "cc_attack_count:" .. key
            local count, _ = redis_util.incr(key, attack_config.duration)
            if not count then
                redis_util.set(key, 1, attack_config.duration)
            elseif count >= attack_config.threshold then
                _M.block_ip(ip, attack_config)
                return
            end
        else
            key = ip .. "attack"
            local limit = ngx.shared.waf_limit
            local count, _ = limit:incr(key, 1, 0, attack_config.duration)

            if not count then
                limit:set(key, 1, attack_config.duration)
            elseif count >= attack_config.threshold then
                _M.block_ip(ip, attack_config)
                return
            end
        end
    end
end

function _M.exec_action(rule_config, match_rule, data)
    local action = rule_config.action

    if match_rule then
        rule_config.rule = match_rule.rule
    else
        rule_config.rule = "默认"
    end

    ngx.ctx.rule_table = rule_config
    ngx.ctx.action = action
    ngx.ctx.hitData = data
    ngx.ctx.is_attack = true

    if rule_config.ipBlock and rule_config.ipBlock == 'on' then
        _M.block_ip(ngx.ctx.ip, rule_config)
    end

    if rule_config.type == nil then
        rule_config.type = "默认"
    end

    attack_count(rule_config.type)

    local msg = "访问 IP " .. ngx.ctx.ip .. " 访问 URL" .. ngx.var.uri .. " 触发动作 " .. action .. " User-Agent " .. ngx.ctx.ua .. "  规则类型 " .. rule_config.type .. "  规则 " .. rule_config.rule

    ngx.log(ngx.ERR, msg)
    if action == "allow" then
        return

    elseif action == "deny" then
        if rule_config.code and rule_config.code ~= 444 then
            deny(rule_config.code, rule_config.res)
        else
            deny(444)
        end

    elseif action == "slide" then
        slide()

    elseif action == "fives" then
        five_second()

    else
        redirect(444)
    end

end

return _M
