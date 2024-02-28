local error = error
local str_len = string.len
local new_table = table.new
local concat_table = table.concat
local insert_table = table.insert
local byte_str = string.byte
local sub_str = string.sub
local type = type
local abs = math.abs
local match_str = string.match
local ngx_re_gsub = ngx.re.gsub

local _M = {}

local INDEX_OUT_OF_RANGE = "String index out of range: "
local NOT_NUMBER = "number expected, got "
local NOT_STRING = "string expected, got "
local NOT_STRING_NIL = "string expected, got nil"

function _M.to_char_array(str)
    local array
    if str then
        local length = str_len(str)
        array = new_table(length, 0)

        local byteLength = 1
        local i, j = 1, 1
        while i <= length do
            local firstByte = byte_str(str, i)
            if firstByte >= 0 and firstByte < 128 then
                byteLength = 1

            elseif firstByte > 191 and firstByte < 224 then
                byteLength = 2

            elseif firstByte > 223 and firstByte < 240 then
                byteLength = 3
                
            elseif firstByte > 239 and firstByte < 248 then
                byteLength = 4
            end

            j = i + byteLength
            local char = sub_str(str, i, j - 1)
            i = j
            insert_table(array, char)
        end
    end

    return array
end

function _M.sub(str, i, j)
    local str_sub
    if str then
        if i == nil then
            i = 1
        end

        if type(i) ~= "number" then
            error(NOT_NUMBER .. type(i))
        end

        if i < 1 then
            error(INDEX_OUT_OF_RANGE .. i)
        end

        if j then
            if type(j) ~= "number" then
                error(NOT_NUMBER .. type(j))
            end
        end

        local array = _M.to_char_array(str)
        if array then
            local length = #array
            local subLen = length - i
            if subLen < 0 then
                error(INDEX_OUT_OF_RANGE .. subLen)
            end

            if not j then
                str_sub = concat_table(array, "", i)
            else
                if abs(j) > length then
                    error(INDEX_OUT_OF_RANGE .. j)
                end
                if j < 0 then
                    j = length + j + 1
                end
                str_sub = concat_table(array, "", i, j)
            end
        end
    end

    return str_sub
end

function _M.trim(str)
    if str then
        str = ngx_re_gsub(str, "^\\s*|\\s*$", "", "jo")
    end

    return str
end

function _M.len(str)
    local str_length = 0
    if str then
        if type(str) ~= "string" then
            error(NOT_STRING .. type(str))
        end

        local length = str_len(str)

        local i = 1
        while i <= length do
            local firstByte = byte_str(str, i)
            if firstByte >= 0 and firstByte < 128 then
                i = i + 1

            elseif firstByte > 191 and firstByte < 224 then
                i = i + 2

            elseif firstByte > 223 and firstByte < 240 then
                i = i + 3

            elseif firstByte > 239 and firstByte < 248 then
                i = i + 4
            end

            str_length = str_length + 1
        end
    else
        error(NOT_STRING_NIL)
    end

    return str_length
end

function _M.default_if_blank(str, default_str)
    if str == nil or match_str(str, "^%s*$") then
        return default_str
    end

    return str
end

return _M
