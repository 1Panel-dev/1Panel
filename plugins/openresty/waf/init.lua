local db = require "db"
local config = require "config"
local mlcache = require "resty.mlcache"

local cache, err = mlcache.new("config", "waf", {
    lru_size = 1000,
    ipc_shm = "ipc_shared_dict",
})
if not cache then
    error("could not create mlcache: " .. err)
end
_G.cache = cache


config.load_config_file()
db.init()





