local cjson = require "cjson"
local pairs = pairs
local insert_table = table.insert
local lower_str = string.lower
local open_file = io.open
local decode = cjson.decode

local _M = {}

function _M.read_rule(file_path, file_name, read_all)
    local file, err = open_file(file_path .. file_name .. ".json", "r")
    if not file then
        ngx.log(ngx.ERR, "Failed to open file ", err)
        return
    end

    local rules_table = {}
    local other_table = {}
    local text = file:read('*a')

    file:close()

    if #text > 0 then
        local result = decode(text)

        if result then
            for key, value in pairs(result) do
                if key == "rules" then
                    for _, r in pairs(value) do
                        if read_all then
                            r.hits = 0
                            r.totalHits = 0
                            insert_table(rules_table, r)
                        else
                            if lower_str(r.state) == 'on' then
                                r.hits = 0
                                r.totalHits = 0
                                insert_table(rules_table, r)
                            end
                        end
                    end
                else
                    other_table[key] = value
                end
            end
        end
    end

    return rules_table, other_table
end

function _M.read_file2table(file_path)
    local file = open_file(file_path, 'r')
    if file == nil then
        return nil
    end
    str = file:read("*a")
    file:close()
    return decode(str)
end

function _M.read_file2string(file_path, binary)
    if not file_path then
        ngx.log(ngx.ERR, "No file found ", file_path)
        return
    end

    local mode = "r"
    if binary == true then
        mode = "rb"
    end

    local file, err = open_file(file_path, mode)
    if not file then
        ngx.log(ngx.ERR, "Failed to open file ", err)
        return
    end

    local content = ""
    repeat
        local chunk = file:read(8192) -- 读取 8KB 的块
        if chunk then
            content = content .. chunk
        else
            break
        end
    until not chunk

    file:close()
    return content
end

return _M