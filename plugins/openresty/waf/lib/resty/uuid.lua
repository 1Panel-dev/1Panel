-- vim:set ts=4 sts=4 sw=4 et:

--- jit-uuid
-- Fast and dependency-free UUID library for LuaJIT/ngx_lua.
-- @module jit-uuid
-- @author Thibault Charbonnier
-- @license MIT
-- @release 0.0.7


local bit = require 'bit'


local tohex = bit.tohex
local band = bit.band
local bor = bit.bor


local _M = {
    _VERSION = '0.0.7'
}


----------
-- seeding
----------


--- Seed the random number generator.
-- Under the hood, this function calls `math.randomseed`.
-- It makes sure to use the most appropriate seeding technique for
-- the current environment, guaranteeing a unique seed.
--
-- To guarantee unique UUIDs, you must have correctly seeded
-- the Lua pseudo-random generator (with `math.randomseed`).
-- You are free to seed it any way you want, but this function
-- can do it for you if you'd like, with some added guarantees.
--
-- @param[type=number] seed (Optional) A seed to use. If none given, will
-- generate one trying to use the most appropriate technique.
-- @treturn number `seed`: the seed given to `math.randomseed`.
-- @usage
-- local uuid = require 'resty.jit-uuid'
-- uuid.seed()
--
-- -- in ngx_lua, seed in the init_worker context:
-- init_worker_by_lua {
--   local uuid = require 'resty.jit-uuid'
--   uuid.seed()
-- }
function _M.seed(seed)
    if not seed then
        if ngx then
            seed = ngx.time() + ngx.worker.pid()

        elseif package.loaded['socket'] and package.loaded['socket'].gettime then
            seed = package.loaded['socket'].gettime()*10000

        else
            seed = os.time()
        end
    end

    math.randomseed(seed)

    return seed
end


-------------
-- validation
-------------


do
    if ngx and string.find(ngx.config.nginx_configure(),'--with-pcre-jit',nil,true) then
        local type = type
        local re_find = ngx.re.find
        local regex = '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'


        --- Validate a string as a UUID.
        -- To be considered valid, a UUID must be given in its canonical
        -- form (hexadecimal digits including the hyphen characters).
        -- This function validates UUIDs disregarding their generation algorithm,
        -- and in a case-insensitive manner, but checks the variant field.
        --
        -- Use JIT PCRE if available in OpenResty or fallbacks on Lua patterns.
        --
        -- @param[type=string] str String to verify.
        -- @treturn boolean `valid`: true if valid UUID, false otherwise.
        -- @usage
        -- local uuid = require 'resty.jit-uuid'
        --
        -- uuid.is_valid 'cbb297c0-a956-486d-ad1d-f9bZZZZZZZZZ' --> false
        -- uuid.is_valid 'cbb297c0-a956-486d-dd1d-f9b42df9465a' --> false (invalid variant)
        -- uuid.is_valid 'cbb297c0a956486dad1df9b42df9465a'     --> false (no dashes)
        -- uuid.is_valid 'cbb297c0-a956-486d-ad1d-f9b42df9465a' --> true
        function _M.is_valid(str)
            -- it has proven itself efficient to first check the length with an
            -- evenly distributed set of valid and invalid uuid lengths.
            if type(str) ~= 'string' or #str ~= 36 then
                return false
            end

            return re_find(str, regex, 'ioj') ~= nil
        end

    else
        local match = string.match
        local d = '[0-9a-fA-F]'
        local p = '^' .. table.concat({
            d:rep(8),
            d:rep(4),
            d:rep(4),
            '[89ab]' .. d:rep(3),
            d:rep(12)
        }, '%-') .. '$'


        function _M.is_valid(str)
            if type(str) ~= 'string' or #str ~= 36 then
                return false
            end

            return match(str, p) ~= nil
        end
    end
end


----------------
-- v4 generation
----------------


do
    local fmt = string.format
    local random = math.random


    --- Generate a v4 UUID.
    -- v4 UUIDs are created from randomly generated numbers.
    --
    -- @treturn string `uuid`: a v4 (randomly generated) UUID.
    -- @usage
    -- local uuid = require 'resty.jit-uuid'
    --
    -- local u1 = uuid()             ---> __call metamethod
    -- local u2 = uuid.generate_v4()
    function _M.generate_v4()
        return (fmt('%s%s%s%s-%s%s-%s%s-%s%s-%s%s%s%s%s%s',
                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),

                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),

                tohex(bor(band(random(0, 255), 0x0F), 0x40), 2),
                tohex(random(0, 255), 2),

                tohex(bor(band(random(0, 255), 0x3F), 0x80), 2),
                tohex(random(0, 255), 2),

                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2),
                tohex(random(0, 255), 2)))
    end
end


----------------
-- v3/v5 generation
----------------


do
    if ngx then
        local ffi = require 'ffi'


        local tonumber = tonumber
        local assert   = assert
        local error    = error
        local concat   = table.concat
        local type     = type
        local char     = string.char
        local fmt      = string.format
        local sub      = string.sub
        local gmatch   = ngx.re.gmatch
        local sha1_bin = ngx.sha1_bin
        local md5      = ngx.md5
        local C        = ffi.C
        local ffi_new  = ffi.new
        local ffi_str  = ffi.string
        local ffi_cast = ffi.cast
        local new_tab
        do
            local ok
            ok, new_tab = pcall(require, 'table.new')
            if not ok then
                new_tab = function(narr, nrec) return {} end
            end
        end


        ffi.cdef [[
            typedef unsigned char u_char;
            typedef intptr_t ngx_int_t;

            u_char * ngx_hex_dump(u_char *dst, const u_char *src, size_t len);
            ngx_int_t ngx_hextoi(u_char *line, size_t n);
        ]]


        local str_type    = ffi.typeof('uint8_t[?]')
        local u_char_type = ffi.typeof('u_char *')


        local function bin_tohex(s)
            local slen = #s
            local blen = slen * 2
            local buf = ffi_new(str_type, blen)

            C.ngx_hex_dump(buf, s, slen)

            return ffi_str(buf, blen)
        end


        local function hex_to_i(s)
            local buf = ffi_cast(u_char_type, s)

            local n = tonumber(C.ngx_hextoi(buf, #s))
            if n == -1 then
                error("could not convert hex to number")
            end

            return n
        end


        local buf = new_tab(16, 0)


        local function factory(namespace, hash_fn)
            if not _M.is_valid(namespace) then
                return nil, 'namespace must be a valid UUID'
            end

            local i = 0
            local iter, err = gmatch(namespace, [[([\da-f][\da-f])]])
            if not iter then
                return nil, 'could not create iter: ' .. err
            end

            while true do
                local m, err = iter()
                if err then
                    return nil, err
                end

                if not m then
                    break
                end

                i = i + 1
                buf[i] = char(tonumber(m[0], 16))
            end

            assert(i == 16, "invalid binary namespace buffer length")
            local ns = concat(buf)

            return function(name)
                if type(name) ~= 'string' then
                    return nil, 'name must be a string'
                end

                local hash, ver, var = hash_fn(ns, name)

                return (fmt('%s-%s-%s%s-%s%s-%s', sub(hash, 1, 8),
                        sub(hash, 9, 12),
                        ver,
                        sub(hash, 15, 16),
                        var,
                        sub(hash, 19, 20),
                        sub(hash, 21, 32)))
            end
        end


        local function v3_hash(binary, name)
            local hash = md5(binary .. name)

            return hash,
            tohex(bor(band(hex_to_i(sub(hash, 13, 14)), 0x0F), 0x30), 2),
            tohex(bor(band(hex_to_i(sub(hash, 17, 18)), 0x3F), 0x80), 2)
        end


        local function v5_hash(binary, name)
            local hash = bin_tohex(sha1_bin(binary .. name))

            return hash,
            tohex(bor(band(hex_to_i(sub(hash, 13, 14)), 0x0F), 0x50), 2),
            tohex(bor(band(hex_to_i(sub(hash, 17, 18)), 0x3F), 0x80), 2)
        end


        --- Instanciate a v3 UUID factory.
        -- @function factory_v3
        -- Creates a closure generating namespaced v3 UUIDs.
        -- @param[type=string] namespace (must be a valid UUID according to `is_valid`)
        -- @treturn function `factory`: a v3 UUID generator.
        -- @treturn string `err`: a string describing an error
        -- @usage
        -- local uuid = require 'resty.jit-uuid'
        --
        -- local fact = assert(uuid.factory_v3('e6ebd542-06ae-11e6-8e82-bba81706b27d'))
        --
        -- local u1 = fact('hello')
        -- ---> 3db7a435-8c56-359d-a563-1b69e6802c78
        --
        -- local u2 = fact('foobar')
        -- ---> e8d3eeba-7723-3b72-bbc5-8f598afa6773
        function _M.factory_v3(namespace)
            return factory(namespace, v3_hash)
        end


        --- Instanciate a v5 UUID factory.
        -- @function factory_v5
        -- Creates a closure generating namespaced v5 UUIDs.
        -- @param[type=string] namespace (must be a valid UUID according to `is_valid`)
        -- @treturn function `factory`: a v5 UUID generator.
        -- @treturn string `err`: a string describing an error
        -- @usage
        -- local uuid = require 'resty.jit-uuid'
        --
        -- local fact = assert(uuid.factory_v5('e6ebd542-06ae-11e6-8e82-bba81706b27d'))
        --
        -- local u1 = fact('hello')
        -- ---> 4850816f-1658-5890-8bfd-1ed14251f1f0
        --
        -- local u2 = fact('foobar')
        -- ---> c9be99fc-326b-5066-bdba-dcd31a6d01ab
        function _M.factory_v5(namespace)
            return factory(namespace, v5_hash)
        end


        --- Generate a v3 UUID.
        -- v3 UUIDs are created from a namespace and a name (a UUID and a string).
        -- The same name and namespace result in the same UUID. The same name and
        -- different namespaces result in different UUIDs, and vice-versa.
        -- The resulting UUID is derived using MD5 hashing.
        --
        -- This is a sugar function which instanciates a short-lived v3 UUID factory.
        -- It is an expensive operation, and intensive generation using the same
        -- namespaces should prefer allocating their own long-lived factory with
        -- `factory_v3`.
        --
        -- @param[type=string] namespace (must be a valid UUID according to `is_valid`)
        -- @param[type=string] name
        -- @treturn string `uuid`: a v3 (namespaced) UUID.
        -- @treturn string `err`: a string describing an error
        -- @usage
        -- local uuid = require 'resty.jit-uuid'
        --
        -- local u = uuid.generate_v3('e6ebd542-06ae-11e6-8e82-bba81706b27d', 'hello')
        -- ---> 3db7a435-8c56-359d-a563-1b69e6802c78
        function _M.generate_v3(namespace, name)
            local fact, err = _M.factory_v3(namespace)
            if not fact then
                return nil, err
            end

            return fact(name)
        end


        --- Generate a v5 UUID.
        -- v5 UUIDs are created from a namespace and a name (a UUID and a string).
        -- The same name and namespace result in the same UUID. The same name and
        -- different namespaces result in different UUIDs, and vice-versa.
        -- The resulting UUID is derived using SHA-1 hashing.
        --
        -- This is a sugar function which instanciates a short-lived v5 UUID factory.
        -- It is an expensive operation, and intensive generation using the same
        -- namespaces should prefer allocating their own long-lived factory with
        -- `factory_v5`.
        --
        -- @param[type=string] namespace (must be a valid UUID according to `is_valid`)
        -- @param[type=string] name
        -- @treturn string `uuid`: a v5 (namespaced) UUID.
        -- @treturn string `err`: a string describing an error
        -- @usage
        -- local uuid = require 'resty.jit-uuid'
        --
        -- local u = uuid.generate_v5('e6ebd542-06ae-11e6-8e82-bba81706b27d', 'hello')
        -- ---> 4850816f-1658-5890-8bfd-1ed14251f1f0
        function _M.generate_v5(namespace, name)
            local fact, err = _M.factory_v5(namespace)
            if not fact then
                return nil, err
            end

            return fact(name)
        end

    else
        function _M.factory_v3() error('v3 UUID generation only supported in ngx_lua', 2) end
        function _M.generate_v3() error('v3 UUID generation only supported in ngx_lua', 2) end
        function _M.factory_v5() error('v5 UUID generation only supported in ngx_lua', 2) end
        function _M.generate_v5() error('v5 UUID generation only supported in ngx_lua', 2) end
    end
end


return setmetatable(_M, {
    __call = _M.generate_v4
})