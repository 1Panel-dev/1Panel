local _M = {}

local bit = require "bit"
local ffi = require "ffi"

local ffi_new = ffi.new
local ffi_string = ffi.string

-- enum sqli_flags
local FLAG_NONE = 0
local FLAG_QUOTE_NONE = 1
local FLAG_QUOTE_SINGLE = 2
local FLAG_QUOTE_DOUBLE = 4
local FLAG_SQL_ANSI = 8
local FLAG_SQL_MYSQL = 16

-- enum lookup_type
local LOOKUP_FINGERPRINT = 4

-- enum html5_flags
local DATA_STATE = 0
local VALUE_NO_QUOTE = 1
local VALUE_SINGLE_QUOTE = 2
local VALUE_DOUBLE_QUOTE = 3
local VALUE_BACK_QUOTE = 4

-- cached ORs
local QUOTE_NONE_SQL_ANSI = bit.bor(FLAG_QUOTE_NONE, FLAG_SQL_ANSI)
local QUOTE_NONE_SQL_MYSQL = bit.bor(FLAG_QUOTE_NONE, FLAG_SQL_MYSQL)
local QUOTE_SINGLE_SQL_ANSI = bit.bor(FLAG_QUOTE_SINGLE, FLAG_SQL_ANSI)
local QUOTE_SINGLE_SQL_MYSQL = bit.bor(FLAG_QUOTE_SINGLE, FLAG_SQL_MYSQL)
local QUOTE_DOUBLE_SQL_MYSQL = bit.bor(FLAG_QUOTE_DOUBLE, FLAG_SQL_MYSQL)

-- libibjection.so
ffi.cdef [[
const char* libinjection_sqli_fingerprint(struct libinjection_sqli_state* sql_state, int flags);

struct libinjection_sqli_token {
	char type;
	char str_open;
	char str_close;
	size_t pos;
	size_t len;
	int count;
	char val[32];
};

typedef char (*ptr_lookup_fn)(struct libinjection_sqli_state*, int lookuptype, const char* word, size_t len);

struct libinjection_sqli_state {
	const char *s;
	size_t slen;
	ptr_lookup_fn lookup;
	void* userdata;
	int flags;
	size_t pos;
	struct libinjection_sqli_token tokenvec[8];
	struct libinjection_sqli_token *current;
	char fingerprint[8];
	int reason;
	int stats_comment_ddw;
	int stats_comment_ddx;
	int stats_comment_c;
	int stats_comment_hash;
	int stats_folds;
	int stats_tokens;
};

void libinjection_sqli_init(struct libinjection_sqli_state * sf, const char *s, size_t len, int flags);
int libinjection_is_sqli(struct libinjection_sqli_state* sql_state);

int libinjection_sqli(const char* s, size_t slen, char fingerprint[]);

int libinjection_is_xss(const char* s, size_t len, int flags);
int libinjection_xss(const char* s, size_t slen);
]]

_M.version = "0.1.1"

local state_type = ffi.typeof("struct libinjection_sqli_state[1]")
local lib, loaded

-- "borrowed" from CF aho-corasick lib
local function _loadlib()
    if (not loaded) then
        local path, so_path
        local libname = "libinjection.so"

        for k, v in string.gmatch(package.cpath, "[^;]+") do
            so_path = string.match(k, "(.*/)")
            if so_path then
                -- "so_path" could be nil. e.g, the dir path component is "."
                so_path = so_path .. libname

                -- Don't get me wrong, the only way to know if a file exist is
                -- trying to open it.
                local f = io.open(so_path)
                if f ~= nil then
                    io.close(f)
                    path = so_path
                    break
                end
            end
        end

        path = "/usr/local/openresty/1pwaf/data/libinjection.so"

        lib = ffi.load(path)

        if (lib) then
            loaded = true
            return true
        else
            return false
        end
    else
        return true
    end
end

-- this function is not publicly exposed so we need to emulate it here. not great but not a measurable perf hit
local function _reparse_as_mysql(sqli_state)
    return sqli_state[0].stats_comment_ddx ~= 0 or sqli_state[0].stats_comment_hash ~= 0
end

--[[
Secondary API: detects SQLi in a string, given a context. Given a string, returns a list of

* boolean indicating a match
* SQLi fingerprint
--]]
local function _sqli_contextwrapper(string, char, flag1, flag2)
    if (char and not string.find(string, char, 1, true)) then
        return false, nil
    end

    if (not loaded) then
        if (not _loadlib()) then
            return false, nil
        end
    end

    local issqli, lookup, sqli_state

    -- allocate a new libinjection_sqli_state struct
    sqli_state = ffi_new(state_type)

    -- init the state
    lib.libinjection_sqli_init(
            sqli_state,
            string,
            #string,
            FLAG_NONE
    )

    -- initial fingerprint
    lib.libinjection_sqli_fingerprint(
            sqli_state,
            flag1
    )

    -- lookup
    lookup = sqli_state[0].lookup(
            sqli_state,
            LOOKUP_FINGERPRINT,
            sqli_state[0].fingerprint,
            #ffi.string(sqli_state[0].fingerprint)
    )

    -- match? great, we're done
    if (lookup > 0) then
        return true, ffi_string(sqli_state[0].fingerprint)
    end

    -- no? reparse, fingerprint and lookup again
    if (flag2 and _reparse_as_mysql(sqli_state)) then
        lib.libinjection_sqli_fingerprint(
                sqli_state,
                flag2
        )

        lookup = sqli_state[0].lookup(
                sqli_state,
                LOOKUP_FINGERPRINT,
                sqli_state[0].fingerprint,
                #ffi.string(sqli_state[0].fingerprint)
        )

        if (lookup > 0) then
            return true, ffi_string(sqli_state[0].fingerprint)
        end
    end

    return false, nil
end

--[[
Wrapper for second-level API with no char context
--]]
function _M.sqli_noquote(string)
    return _sqli_contextwrapper(
            string,
            nil,
            QUOTE_NONE_SQL_ANSI,
            QUOTE_NONE_SQL_MYSQL
    )
end

--[[
Wrapper for second-level API with CHAR_SINGLE context
--]]
function _M.sqli_singlequote(string)
    return _sqli_contextwrapper(
            string,
            "'",
            QUOTE_SINGLE_SQL_ANSI,
            QUOTE_SINGLE_SQL_MYSQL
    )
end

--[[
Wrapper for second-level API with CHAR_DOUBLE context
--]]
function _M.sqli_doublequote(string)
    return _sqli_contextwrapper(
            string,
            '"',
            QUOTE_DOUBLE_SQL_MYSQL
    )
end

--[[
Simple API. Given a string, returns a list of

* boolean indicating a match
* SQLi fingerprint
--]]
function _M.sqli(string)
    if (not loaded) then
        if (not _loadlib()) then
            return false, nil
        end
    end

    local fingerprint = ffi_new("char [8]")

    return lib.libinjection_sqli(string, #string, fingerprint) == 1, ffi_string(fingerprint)
end

--[[
Secondary API: detects XSS in a string, given a context. Given a string, returns a boolean denoting if XSS was detected
--]]
local function _xss_contextwrapper(string, flag)
    if (not loaded) then
        if (not _loadlib()) then
            return false
        end
    end

    return lib.libinjection_is_xss(string, #string, flag) == 1
end

--[[
Wrapper for second-level API with DATA_STATE flag
--]]
function _M.xss_data_state(string)
    return _xss_contextwrapper(
            string,
            DATA_STATE
    )
end

--[[
Wrapper for second-level API with VALUE_NO_QUOTE flag
--]]
function _M.xss_noquote(string)
    return _xss_contextwrapper(
            string,
            VALUE_NO_QUOTE
    )
end

--[[
Wrapper for second-level API with VALUE_SINGLE_QUOTE flag
--]]
function _M.xss_singlequote(string)
    return _xss_contextwrapper(
            string,
            VALUE_SINGLE_QUOTE
    )
end

--[[
Wrapper for second-level API with VALUE_DOUBLE_QUOTE flag
--]]
function _M.xss_doublequote(string)
    return _xss_contextwrapper(
            string,
            VALUE_DOUBLE_QUOTE
    )
end

--[[
Wrapper for second-level API with VALUE_BACK_QUOTE flag
--]]
function _M.xss_backquote(string)
    return _xss_contextwrapper(
            string,
            VALUE_BACK_QUOTE
    )
end

--[[
ALPHA version of XSS detector. Given a string, returns a boolean denoting if XSS was detected
--]]
function _M.xss(string)
    if (not loaded) then
        if (not _loadlib()) then
            return false
        end
    end

    return lib.libinjection_xss(string, #string) == 1
end

return _M