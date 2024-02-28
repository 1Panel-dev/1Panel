local logger = require "logger"

local loggers = {}

local _M = {}

function _M.get_logger(log_path, host, rolling)
    local host_logger = loggers[host]
    if not host_logger then
        host_logger = logger:new(log_path, host, rolling)
        loggers[host] = host_logger
    end
    return host_logger
end

return _M