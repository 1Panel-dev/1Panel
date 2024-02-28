local config = require "config"
local redis_util = require "redis_util"
local utils = require "utils"

local _M = {}

function _M.set_access_token(k, v)
    local secret = config.get_secret()
    local key = ngx.md5(ngx.ctx.ip .. ngx.var.server_name .. ngx.ctx.website_key
            .. ngx.ctx.ua .. ngx.ctx.today .. secret)
    local value = ngx.md5(ngx.time() .. ngx.ctx.ip)
    --TODO check value
    if key ~= k then
        ngx.exit(444)
    end
    ngx.log(ngx.ERR, "set cc key: ", key)
    if config.redis_on then
        --local prefix = "ac_token:"
        --redis_util.set(prefix .. accesstoken, accesstoken, timeout)
    else
        local limit = ngx.shared.waf_accesstoken
        limit:set(key, value, 7200)
    end

    local cookie_expire = ngx.cookie_time(ngx.time() + 86400)
    ngx.header['Set-Cookie'] = { key .. '=' .. value .. '; path=/; Expires=' .. cookie_expire }
    ngx.exit(200)
end

function _M.check_access_token()
    local secret = config.get_secret()
    local key = ngx.md5(ngx.ctx.ip .. ngx.var.server_name .. ngx.ctx.website_key
            .. ngx.ctx.ua .. ngx.ctx.today .. secret)
    if not ngx.var.http_cookie then
        return false
    end
    local cookies = utils.get_cookie_list(ngx.var.http_cookie)
    if not cookies then
        return false
    end
    if not cookies[key] then
        return false
    end
    local accesstoken = cookies[key]
    local value = nil

    if config.redis_on then
        local prefix = "ac_token:"
        value = redis_util.get(prefix .. key)
        if value and value == accesstoken then
            return true
        end
    else
        local limit = ngx.shared.waf_accesstoken
        value = limit:get(key)
    end
    if value and value == accesstoken then
        return true
    end
    return false
end

function _M.clear_access_token()
    ngx.header['Set-Cookie'] = { 'a_token=; path=/; Expires=Thu, 01-Jan-1970 00:00:00 GMT' }
end

return _M
