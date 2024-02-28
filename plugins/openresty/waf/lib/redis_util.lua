local redis = require "resty.redis"
local config = require "config"

local _M = {}

local connect_timeout, send_timeout, read_timeout = 1000, 1000, 1000

function _M.get_conn()
    local red, err1 = redis:new()
    if not red then
        ngx.log(ngx.ERR, "failed to new redis:", err1)
        return nil, err1
    end

    local redis_config = config.get_redis_config()

    red:set_timeouts(connect_timeout, send_timeout, read_timeout)

    local ok, err = red:connect(redis_config.host, redis_config.port, { ssl = redis_config.ssl, pool_size = redis_config.poolSize })

    if not ok then
        ngx.log(ngx.ERR, "failed to connect: ", err .. "\n")
        return nil, err
    end

    if redis_config.password ~= nil and #redis_config.password ~= 0 then
        local times = 0
        times, err = red:get_reused_times()

        if times == 0 then
            local res, err2 = red:auth(redis_config.password)
            if not res then
                ngx.log(ngx.ERR, "failed to authenticate: ", err2)
                return nil, err2
            end
        end
    end

    return red, err
end

function _M.close_conn(red)
    local ok, err = red:set_keepalive(10000, 100)
    if not ok then
        ngx.log(ngx.ERR, "failed to set keepalive: ", err)
    end

    return ok, err
end

function _M.set(key, value, expire_time)
    local red, _ = _M.get_conn()
    local ok, err = nil, nil
    if red then
        ok, err = red:set(key, value)
        if not ok then
            ngx.log(ngx.ERR, "failed to set key: " .. key .. " ", err)
        elseif expire_time and expire_time > 0 then
            red:expire(key, expire_time)
        end

        _M.close_conn(red)
    end

    return ok, err
end

function _M.bath_set(keyTable, value, keyPrefix)
    local red, _ = _M.get_conn()
    local results, err = nil, nil
    if red then
        red:init_pipeline()

        if keyPrefix then
            for _, ip in ipairs(keyTable) do
                red:set(keyPrefix .. ip, value)
            end
        else
            for _, ip in ipairs(keyTable) do
                red:set(ip, value)
            end
        end

        results, err = red:commit_pipeline()
        if not results then
            ngx.log(ngx.ERR, "failed to set keys: ", err)
        end

        _M.close_conn(red)
    end

    return results, err
end

function _M.get(key)
    local red, err = _M.get_conn()
    local value = nil
    if red then
        value, err = red:get(key)
        if not value then
            ngx.log(ngx.ERR, "failed to get key: " .. key, err)
            return value, err
        end
        if value == ngx.null then
            value = nil
        end

        _M.close_conn(red)
    end

    return value, err
end

function _M.incr(key, expire_time)
    local red, err = _M.get_conn()
    local res = nil
    if red then
        res, err = red:incr(key)
        if not res then
            ngx.log(ngx.ERR, "failed to incr key: " .. key, err)
        elseif res == 1 and expire_time and expire_time > 0 then
            red:expire(key, expire_time)
        end

        _M.close_conn(red)
    end

    return res, err
end

return _M
