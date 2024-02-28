function ip_to_int(ip)
    local ip_int = 0
    for i, octet in ipairs({ ip:match("(%d+)%.(%d+)%.(%d+)%.(%d+)") }) do
        ip_int = ip_int + tonumber(octet) * 256 ^ (4 - i)
    end
    return ip_int
end

------ 示例
local ip_address = "222.249.139.98"
local ip_number = ip_to_int(ip_address)
print(ip_number)

--local geoip = require "lib.resty.maxminddb"
--local cjson = require("cjson")
--
--geoip.init("/Users/wangzhengkun/Downloads/blackIP.mmdb")
--
--local geo = geoip.lookup("165.154.132.251")
--
--print(cjson.encode(geo))

--local fileUtils = require "lib.file"
--local read_file2string = fileUtils.read_file2string
--
--local slideHtml = read_file2string("./html/" .. "slide.html")
--
--print(string.format(slideHtml, "1", "2"))


--local today = os.date("%Y-%m-%d")
--print(today)

