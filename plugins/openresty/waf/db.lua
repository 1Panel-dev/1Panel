local config = require "config"

local open_file = io.open
local exec = os.execute
local pcall = pcall

local _M = {}

local function init_dir(path)
    local file = open_file(path, "rb")
    if not file then
        exec("mkdir -p " .. path)
    end
end

local function check_table(table_name)
    if wafdb == nil then
        return false
    end
    local stmt = wafdb:prepare("SELECT COUNT(*) FROM sqlite_master where type='table' and name=?")
    local rows = 0
    if stmt ~= nil then
        stmt:bind_values(table_name)
        stmt:step()
        rows = stmt:get_uvalues()
        stmt:finalize()
    end
    return rows > 0
end

function _M.init_db()
    local ok, sqlite3 = pcall(function()
        return require "lsqlite3"
    end)
    if not ok then
        return false
    end
    if wafdb then
        return false
    end
    local path = config.waf_dir .. "db/"
    init_dir(path)
    local db_path = path .. "1pwaf.db"
    if wafdb == nil or not wafdb:isopen() then
        wafdb = sqlite3.open(db_path)
        if wafdb == nil then
            return false
        end
        wafdb:exec([[PRAGMA journal_mode = wal]])
        wafdb:exec([[PRAGMA synchronous = 0]])
        wafdb:exec([[PRAGMA page_size = 8192]])
        wafdb:exec([[PRAGMA journal_size_limit = 2147483648]])
    end
    local status = {}
    if not check_table("attack_log") then
        status = wafdb:exec([[
            CREATE TABLE attack_log (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                ip TEXT,
                ip_city TEXT,
                ip_country TEXT,
                ip_subdivisions TEXT,
                ip_continent TEXT,
                ip_longitude TEXT,
                ip_latitude TEXT,
                time INTEGER,
                localtime TEXT,
                server_name TEXT,
                website_key TEXT,
                host TEXT,
                method TEXT,
                uri TEXT,
                user_agent TEXT,
                rule TEXT,
                nginx_log TEXT,
                blocking_time INTEGER,
                action TEXT,
                msg TEXT,
                params TEXT,
                is_block INTEGER
            )]])

        ngx.log(ngx.ERR, "init db status" .. status)
    end

    ngx.log(ngx.ERR, "init db success")
end

return _M