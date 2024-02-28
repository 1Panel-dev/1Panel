-- Copyright (C) 2013-2016 Jiale Zhi (calio), CloudFlare Inc.
-- See RFC6265 http://tools.ietf.org/search/rfc6265
-- require "luacov"

local type          = type
local byte          = string.byte
local sub           = string.sub
local format        = string.format
local log           = ngx.log
local ERR           = ngx.ERR
local WARN          = ngx.WARN
local ngx_header    = ngx.header

local EQUAL         = byte("=")
local SEMICOLON     = byte(";")
local SPACE         = byte(" ")
local HTAB          = byte("\t")

-- table.new(narr, nrec)
local ok, new_tab = pcall(require, "table.new")
if not ok then
    new_tab = function () return {} end
end

local ok, clear_tab = pcall(require, "table.clear")
if not ok then
    clear_tab = function(tab) for k, _ in pairs(tab) do tab[k] = nil end end
end

local _M = new_tab(0, 2)

_M._VERSION = '0.01'


local function get_cookie_table(text_cookie)
    if type(text_cookie) ~= "string" then
        log(ERR, format("expect text_cookie to be \"string\" but found %s",
                type(text_cookie)))
        return {}
    end

    local EXPECT_KEY    = 1
    local EXPECT_VALUE  = 2
    local EXPECT_SP     = 3

    local n = 0
    local len = #text_cookie

    for i=1, len do
        if byte(text_cookie, i) == SEMICOLON then
            n = n + 1
        end
    end

    local cookie_table  = new_tab(0, n + 1)

    local state = EXPECT_SP
    local i = 1
    local j = 1
    local key, value

    while j <= len do
        if state == EXPECT_KEY then
            if byte(text_cookie, j) == EQUAL then
                key = sub(text_cookie, i, j - 1)
                state = EXPECT_VALUE
                i = j + 1
            end
        elseif state == EXPECT_VALUE then
            if byte(text_cookie, j) == SEMICOLON
                    or byte(text_cookie, j) == SPACE
                    or byte(text_cookie, j) == HTAB
            then
                value = sub(text_cookie, i, j - 1)
                cookie_table[key] = value

                key, value = nil, nil
                state = EXPECT_SP
                i = j + 1
            end
        elseif state == EXPECT_SP then
            if byte(text_cookie, j) ~= SPACE
                and byte(text_cookie, j) ~= HTAB
            then
                state = EXPECT_KEY
                i = j
                j = j - 1
            end
        end
        j = j + 1
    end

    if key ~= nil and value == nil then
        cookie_table[key] = sub(text_cookie, i)
    end

    return cookie_table
end

function _M.new(self)
    local _cookie = ngx.var.http_cookie
    --if not _cookie then
        --return nil, "no cookie found in current request"
    --end
    return setmetatable({ _cookie = _cookie, set_cookie_table = new_tab(4, 0) },
        { __index = self })
end

function _M.get(self, key)
    if not self._cookie then
        return nil, "no cookie found in the current request"
    end
    if self.cookie_table == nil then
        self.cookie_table = get_cookie_table(self._cookie)
    end

    return self.cookie_table[key]
end

function _M.get_all(self)
    if not self._cookie then
        return nil, "no cookie found in the current request"
    end

    if self.cookie_table == nil then
        self.cookie_table = get_cookie_table(self._cookie)
    end

    return self.cookie_table
end

function _M.get_cookie_size(self)
    if not self._cookie then
        return 0
    end

    return string.len(self._cookie)
end

local function bake(cookie)
    if not cookie.key or not cookie.value then
        return nil, 'missing cookie field "key" or "value"'
    end

    if cookie["max-age"] then
        cookie.max_age = cookie["max-age"]
    end

    if (cookie.samesite) then
        local samesite = cookie.samesite

        -- if we don't have a valid-looking attribute, ignore the attribute
        if (samesite ~= "Strict" and samesite ~= "Lax" and samesite ~= "None") then
            log(WARN, "SameSite value must be 'Strict', 'Lax' or 'None'")
            cookie.samesite = nil
        end
    end

    local str = cookie.key .. "=" .. cookie.value
        .. (cookie.expires and "; Expires=" .. cookie.expires or "")
        .. (cookie.max_age and "; Max-Age=" .. cookie.max_age or "")
        .. (cookie.domain and "; Domain=" .. cookie.domain or "")
        .. (cookie.path and "; Path=" .. cookie.path or "")
        .. (cookie.secure and "; Secure" or "")
        .. (cookie.httponly and "; HttpOnly" or "")
        .. (cookie.samesite and "; SameSite=" .. cookie.samesite or "")
        .. (cookie.extension and "; " .. cookie.extension or "")
    return str
end

function _M.set(self, cookie)
    local cookie_str, err = bake(cookie)
    if not cookie_str then
        return nil, err
    end

    local set_cookie = ngx_header['Set-Cookie']
    local set_cookie_type = type(set_cookie)
    local t = self.set_cookie_table
    clear_tab(t)

    if set_cookie_type == "string" then
        -- only one cookie has been setted
        if set_cookie ~= cookie_str then
            t[1] = set_cookie
            t[2] = cookie_str
            ngx_header['Set-Cookie'] = t
        end
    elseif set_cookie_type == "table" then
        -- more than one cookies has been setted
        local size = #set_cookie

        -- we can not set cookie like ngx.header['Set-Cookie'][3] = val
        -- so create a new table, copy all the values, and then set it back
        for i=1, size do
            t[i] = ngx_header['Set-Cookie'][i]
            if t[i] == cookie_str then
                -- new cookie is duplicated
                return true
            end
        end
        t[size + 1] = cookie_str
        ngx_header['Set-Cookie'] = t
    else
        -- no cookie has been setted
        ngx_header['Set-Cookie'] = cookie_str
    end
    return true
end

_M.get_cookie_string = bake

return _M
