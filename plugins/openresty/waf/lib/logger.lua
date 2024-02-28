local type = type
local concat_table = table.concat
local new_table = table.new
local tostring = tostring
local setmetatable = setmetatable
local open_file = io.open
local ngx_timer_at = ngx.timer.at

local _M = {}

local mt = { __index = _M }

function _M:new(log_path, host, rolling)
    local t = {
        flush_limit = 4096, -- 4kb
        flush_timeout = 1,

        buffered_size = 0,
        buffer_index = 0,
        buffer_data = new_table(20000, 0),

        log_path = log_path,
        prefix = log_path .. host .. '_',
        rolling = rolling or false,
        host = host,
        timer = nil }

    setmetatable(t, mt)
    return t
end

local function needFlush(self)
    if self.buffered_size > 0 then
        return true
    end

    return false
end

local function flush_lock(self)
    local dic_lock = ngx.shared.dict_locks
    local locked = dic_lock:get(self.host)
    if not locked then
        local succ, err = dic_lock:set(self.host, true)
        if not succ then
            ngx.log(ngx.ERR, "failed to lock logfile " .. self.host .. ": ", err)
        end
        return succ
    end
    return false
end

local function flush_unlock(self)
    local dic_lock = ngx.shared.dict_locks
    local success, err = dic_lock:set(self.host, false)
    if not success then
        ngx.log(ngx.ERR, "failed to unlock logfile " .. self.host .. ": ", err)
    end
    return success
end

local function write_file(self, value)
    local file_name = ''
    if self.rolling then
        file_name = self.prefix .. ngx.today() .. ".log"
    else
        file_name = self.log_path
    end

    local file = open_file(file_name, "a+")

    if file == nil or value == nil then
        return
    end

    file:write(value)
    file:flush()
    file:close()

    return
end

local function flushBuffer(self)
    if not needFlush(self) then
        return true
    end

    if not flush_lock(self) then
        return true
    end

    local buffer = concat_table(self.buffer_data, "", 1, self.buffer_index)
    write_file(self, buffer)

    self.buffered_size = 0
    self.buffer_index = 0
    self.buffer_data = new_table(20000, 0)

    flush_unlock(self)
end

local function flushPeriod(premature, self)
    flushBuffer(self)
    self.timer = false
end

local function writeBuffer(self, msg, msg_len)
    self.buffer_index = self.buffer_index + 1

    self.buffer_data[self.buffer_index] = msg

    self.buffered_size = self.buffered_size + msg_len

    return self.buffered_size
end

local function startTimer(self)
    if not self.timer then
        local ok, err = ngx_timer_at(self.flush_timeout, flushPeriod, self)
        if not ok then
            ngx.log(ngx.ERR, "failed to create the timer: ", err)
            return
        end
        if ok then
            self.timer = true
        end
    end
    return self.timer
end

function _M:log(msg)
    if type(msg) ~= "string" then
        msg = tostring(msg)
    end

    local msg_len = #msg
    local len = msg_len + self.buffered_size

    if len < self.flush_limit then
        writeBuffer(self, msg, msg_len)
        startTimer(self)
    elseif len >= self.flush_limit then
        writeBuffer(self, msg, msg_len)
        flushBuffer(self)
    end
end

return _M