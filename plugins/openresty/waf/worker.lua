local uuid = require 'resty.uuid'
local utils = require "utils"
local config = require "config"

uuid.seed()

local update_req_count = function()
    local req_count =  ngx.shared.waf_req_count
    local req_count_update = req_count:get("req_count") or 0
    req_count:set("req_count", 0)
    local count_4xx_update = req_count:get("count_4xx") or 0
    req_count:set("count_4xx", 0)
    local count_5xx_update = req_count:get("count_5xx") or 0
    req_count:set("count_5xx", 0)
    local attack_count_update = req_count:get("attack_count") or 0
    req_count:set("attack_count", 0)
    
    if req_count_update == 0 and count_4xx_update == 0 and count_5xx_update == 0 and attack_count_update == 0 then
        return
    end
    
    local today = ngx.today()
    local wafdb = utils.get_wafdb(config.waf_db_path)
    if not wafdb then
        ngx.log(ngx.ERR, "get log db failed")
        return
    end

    wafdb:execute([[BEGIN TRANSACTION]])

    local stmt_exist = wafdb:prepare("SELECT COUNT(*) FROM waf_stat WHERE day = ?")
    stmt_exist:bind_values(today)
    stmt_exist:step()
    local count = stmt_exist:get_uvalues()
    stmt_exist:finalize()

    local code = 0
    if count > 0 then
        local stmt = wafdb:prepare("UPDATE waf_stat SET req_count = req_count + ?, count4xx = count4xx + ?, count5xx = count5xx + ?, attack_count = attack_count + ? WHERE day = ?")
        stmt:bind_values(req_count_update, count_4xx_update, count_5xx_update, attack_count_update, today)
        code = stmt:step()
        stmt:finalize()
    else
        local stmt = wafdb:prepare("INSERT INTO waf_stat (day, req_count, count4xx, count5xx, attack_count,create_date) VALUES (?, ?, ?, ?, ?,DATETIME('now'))")
        stmt:bind_values(today, req_count_update, count_4xx_update, count_5xx_update, attack_count_update)
        code = stmt:step()
        stmt:finalize()
    end

    wafdb:execute([[COMMIT]])

    --local error_msg = wafdb:errmsg()
    --if error_msg then
    --    ngx.log(ngx.ERR, "update waf_stat error ", error_msg .. "  ")
    --end
end

if 0 == ngx.worker.id() then
    local ok, err = ngx.timer.every(2, update_req_count)
    if not ok then
           ngx.log(ngx.ERR, "failed to create the timer: ", err)
        return
    end
end