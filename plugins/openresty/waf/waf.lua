local geoip = require "geoip"
local lib = require "lib"
local file_utils = require "file"
local config = require "config"
local cc = require "cc"
local utils = require "utils"
local cjson = require "cjson"

local ipairs = ipairs
local sub_str = string.sub
local find_str = string.find
local split_str = utils.split
local encode = cjson.encode
local read_file2table = file_utils.read_file2table
local tonumber = tonumber
local date = os.date
local format_str = string.format

local function get_website_key()
    local s_name = ngx.var.server_name
    local website_key = ngx.shared.waf:get(s_name)
    if website_key then
        return website_key
    end
    local websites = read_file2table(config.config_dir .. '/websites.json')
    if not websites then
        return s_name
    end
    for _, v in ipairs(websites)
    do
        for _, domain in ipairs(v['domains'])
        do
            if s_name == domain then
                ngx.shared.waf:set(s_name, v['key'], 3600)
                return v['key']
            end
        end
    end
    if s_name == '_' then
        s_name = "unknown"
    end
    return s_name
end



local function init()
    local ip = utils.get_real_ip()
    ngx.ctx.ip = ip
    local ua = utils.get_header("user-agent")
    if not ua then
        ua = ""
    end
    
    ngx.ctx.ua = ua
    ngx.ctx.geoip = utils.get_geo_ip(ip)
    
    ngx.ctx.website_key = get_website_key()
    ngx.ctx.method = ngx.req.get_method()
    ngx.ctx.content_type = utils.get_header("content-type")
    if ngx.ctx.content_type then
        ngx.ctx.content_length = tonumber(utils.get_header("content-length"))
    end
    ngx.ctx.today = date("%Y-%m-%d")
end

local function return_js(js_type)
    ngx.header.content_type = "text/html;charset=utf8"
    ngx.header.Cache_Control = "no-cache"
    local host = ngx.var.scheme .. "://" .. ngx.var.host
    local set_access_url = host .. "/set_access_token"
    local secret = config.get_secret()
    local key = ngx.md5(ngx.ctx.ip .. ngx.var.server_name .. ngx.ctx.website_key
            .. ngx.ctx.ua .. ngx.ctx.today .. secret)
    local value = ngx.md5(ngx.time() .. ngx.ctx.ip)
    local js = config.get_html_res(js_type)
    ngx.say(format_str(js, set_access_url, key, value))
    ngx.status = 200
    ngx.exit(200)
end

local function return_json(data)
    ngx.header.content_type = "application/json;"
    ngx.header.Cache_Control = "no-cache"
    ngx.status = 200
    ngx.say(data)
    ngx.exit(200)
end

local function waf_api()
    local uri = ngx.var.uri
    local prefix = sub_str(uri, 1, 15)
    if find_str(prefix, "/set_access_token") then
        local kvs = split_str(uri, "-")
        if kvs[2] and kvs[3] then
            cc.set_access_token(kvs[2], kvs[3])
        else
            ngx.exit(444)
        end
    end
    if uri == "/slide_check_" .. ngx.md5(ngx.ctx.ip) .. ".js" then
        return_js("slide_js")
    end

    if uri == "/5s_check_" .. ngx.md5(ngx.ctx.ip) .. ".js" then
        return_js("five_second_js")
    end

    if ngx.var.remote_addr ~= '127.0.0.1' then
        return false
    end
    if uri == '/reload_waf_config' then
        config.load_config_file()
        ngx.exit(200)
    end
    if uri == '/get_black_ip' then
        --TODO 从 redis 获取黑名单
        local data = ngx.shared.waf_black_ip:get_keys(0)
        return_json(encode(data))
    end
end

if config.is_waf_on() then
    init()
    waf_api()
    
    if ngx.ctx.website_key == "unknown" then
        ngx.exit(403)
        return
    end
    
    if lib.is_white_ip() then
        return true
    end
    lib.default_ip_black()
    lib.black_ip()

    if lib.is_white_ua() then
        return true
    end
    lib.default_ua_black()
    lib.black_ua()

    lib.cc_url()
    if lib.is_white_url() then
        return true
    end
    lib.black_url()

    lib.allow_location_check()
    lib.acl()
    lib.bot_check()
    lib.method_check()
    lib.cc()
    lib.args_check()
    lib.cookie_check()
    lib.post_check()
    lib.header_check()
end