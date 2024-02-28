local db = require "db"
local config = require "config"

config.load_config_file()

db.init_db()



