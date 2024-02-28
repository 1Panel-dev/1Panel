local geo = require "resty.maxminddb"

local pcall = pcall

local _M = {}
local geo_ip_file = "/usr/local/openresty/1pwaf/data/GeoIP.mmdb"
local black_ip_file = "/usr/local/openresty/1pwaf/data/BlackIP.mmdb"

function _M.init()
    if not geo.initted() then
        geo.init({
            geo_ip = geo_ip_file,
            black_ip = black_ip_file
        })
    end
end

function _M.is_default_black_ip(ip)
    local pass, res, err = pcall(geo.lookup, "black_ip", ip)
    if not pass then
        ngx.log(ngx.ERR, 'failed to lookup black ip,reason:', err)
    elseif res and res['isBlack'] then
        return true
    end
    return false
end

function _M.lookup(ip)
    local geo_res = {
        iso = "",
        country = "",
        city = "",
        longitude = 0,
        latitude = 0,
        province = ""
    }
    local pass, res, err = pcall(geo.lookup, "geo_ip", ip)
    if not pass then
        ngx.log(ngx.ERR, 'failed to lookup by ip,reason:', err)
    elseif res and res['iso'] then
        geo_res.iso = res['iso']
        geo_res.country = res['country']
        geo_res.province = res['province']
        geo_res.longitude = res['longitude']
        geo_res.latitude = res['latitude']
        return geo_res
    end
    return geo_res
end

return _M
