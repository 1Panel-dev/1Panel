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

local function check_table(table_name,wafdb)
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

local function init_db_config(db_path)
    local ok, sqlite3 = pcall(function()
        return require "lsqlite3"
    end)
    if not ok then
        return false
    end
    local wafdb = sqlite3.open(db_path)
    if wafdb == nil then
        return false
    end
    wafdb:exec([[PRAGMA journal_mode = wal]])
    wafdb:exec([[PRAGMA synchronous = OFF]])
    wafdb:exec([[PRAGMA page_size = 8192]])
    wafdb:exec([[PRAGMA journal_size_limit = 2147483648]])
    return wafdb
end

function _M.init()
    init_dir(config.waf_db_dir)
    local  wafdb = init_db_config(config.waf_db_path)
    if not wafdb then
        return false
    end

    local status = {}
    if not check_table("waf_stat",wafdb) then
        status = wafdb:exec([[
            CREATE TABLE waf_stat (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                day TEXT,
                req_count INTEGER,
                attack_count INTEGER,
                count4xx INTEGER,
                count5xx INTEGER,
                create_date DATETIME
            )]])
        ngx.log(ngx.ERR, "init waf_stat status"..status)
    end

    local logdb = init_db_config(config.waf_log_db_path)
    if not check_table("req_logs",logdb) then
        status = logdb:exec([[
            CREATE TABLE req_logs (
                id TEXT PRIMARY KEY,
                ip TEXT,
                ip_iso TEXT,
                ip_country_zh TEXT,
                ip_country_en TEXT,
                ip_province_zh TEXT,
                ip_province_en TEXT,
                ip_longitude TEXT,
                ip_latitude TEXT,
                localtime DATETIME,
                server_name TEXT,
                website_key TEXT,
                host TEXT,
                method TEXT,
                uri TEXT,
                user_agent TEXT,
                exec_rule TEXT,
                rule_type TEXT,
                match_rule TEXT,
                match_value TEXT,
                nginx_log TEXT,
                blocking_time INTEGER,
                action TEXT,
                is_block INTEGER,
                is_attack INTEGER
            )]])
    end

    if not check_table("block_ips",logdb) then
        status = logdb:exec([[
            CREATE TABLE block_ips (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                ip TEXT,
                is_block INTEGER,
                blocking_time INTEGER,
                req_log_id TEXT,
                create_date DATETIME
            )]])
        ngx.log(ngx.ERR, "init block_ip status"..status)
    end
    
    ngx.log(ngx.ERR, "init db success")
end

return _M