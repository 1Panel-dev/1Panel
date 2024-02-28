--[[
	Copyright 2017-now anjia (anjia0532@gmail.com)

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at
		http://www.apache.org/licenses/LICENSE-2.0
	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
]]

-- copy from https://github.com/lilien1010/lua-resty-maxminddb/blob/f96633e2428f8f7bcc1e2a7a28b747b33233a8db/resty/maxminddb.lua#L5-L12

local ffi = require('ffi')
local ffi_new = ffi.new
local ffi_str = ffi.string
local ffi_cast = ffi.cast
local ffi_gc = ffi.gc
local C = ffi.C

local _M = {}
local _D = {}

_M._VERSION = '1.3.3'
local mt = { __index = _M }

-- copy from https://github.com/lilien1010/lua-resty-maxminddb/blob/f96633e2428f8f7bcc1e2a7a28b747b33233a8db/resty/maxminddb.lua#L36-L126
ffi.cdef [[

typedef long int ssize_t;

typedef unsigned int mmdb_uint128_t __attribute__ ((__mode__(TI)));

typedef struct MMDB_entry_s {
  struct MMDB_s *mmdb;
  uint32_t offset;
} MMDB_entry_s;

typedef struct MMDB_lookup_result_s {
  bool found_entry;
  MMDB_entry_s entry;
  uint16_t netmask;
} MMDB_lookup_result_s;

typedef struct MMDB_entry_data_s {
  bool has_data;
  union {
    uint32_t pointer;
    const char *utf8_string;
    double double_value;
    const uint8_t *bytes;
    uint16_t uint16;
    uint32_t uint32;
    int32_t int32;
    uint64_t uint64;
    mmdb_uint128_t uint128;
    bool boolean;
    float float_value;
  };

  uint32_t offset;
  uint32_t offset_to_next;
  uint32_t data_size;
  uint32_t type;
} MMDB_entry_data_s;

typedef struct MMDB_entry_data_list_s {
  MMDB_entry_data_s entry_data;
  struct MMDB_entry_data_list_s *next;
} MMDB_entry_data_list_s;

typedef struct MMDB_description_s {
  const char *language;
  const char *description;
} MMDB_description_s;

typedef struct MMDB_metadata_s {
  uint32_t node_count;
  uint16_t record_size;
  uint16_t ip_version;
  const char *database_type;
  struct {
    size_t count;
    const char **names;
  } languages;
  uint16_t binary_format_major_version;
  uint16_t binary_format_minor_version;
  uint64_t build_epoch;
  struct {
    size_t count;
    MMDB_description_s **descriptions;
  } description;
} MMDB_metadata_s;

typedef struct MMDB_ipv4_start_node_s {
  uint16_t netmask;
  uint32_t node_value;
} MMDB_ipv4_start_node_s;

typedef struct MMDB_s {
  uint32_t flags;
  const char *filename;
  ssize_t file_size;
  const uint8_t *file_content;
  const uint8_t *data_section;
  uint32_t data_section_size;
  const uint8_t *metadata_section;
  uint32_t metadata_section_size;
  uint16_t full_record_byte_size;
  uint16_t depth;
  MMDB_ipv4_start_node_s ipv4_start_node;
  MMDB_metadata_s metadata;
} MMDB_s;

typedef  char * pchar;

MMDB_lookup_result_s MMDB_lookup_string(MMDB_s *const mmdb,   const char *const ipstr, int *const gai_error,int *const mmdb_error);
int MMDB_open(const char *const filename, uint32_t flags, MMDB_s *const mmdb);
int MMDB_aget_value(MMDB_entry_s *const start,  MMDB_entry_data_s *const entry_data,  const char *const *const path);
char *MMDB_strerror(int error_code);

int MMDB_get_entry_data_list(MMDB_entry_s *start, MMDB_entry_data_list_s **const entry_data_list);
void MMDB_free_entry_data_list(MMDB_entry_data_list_s *const entry_data_list);
void MMDB_close(MMDB_s *const mmdb);
const char *gai_strerror(int errcode);
]]

-- error codes
-- https://github.com/maxmind/libmaxminddb/blob/master/include/maxminddb.h#L66
local MMDB_SUCCESS = 0
local MMDB_FILE_OPEN_ERROR = 1
local MMDB_CORRUPT_SEARCH_TREE_ERROR = 2
local MMDB_INVALID_METADATA_ERROR = 3
local MMDB_IO_ERROR = 4
local MMDB_OUT_OF_MEMORY_ERROR = 5
local MMDB_UNKNOWN_DATABASE_FORMAT_ERROR = 6
local MMDB_INVALID_DATA_ERROR = 7
local MMDB_INVALID_LOOKUP_PATH_ERROR = 8
local MMDB_LOOKUP_PATH_DOES_NOT_MATCH_DATA_ERROR = 9
local MMDB_INVALID_NODE_NUMBER_ERROR = 10
local MMDB_IPV6_LOOKUP_IN_IPV4_DATABASE_ERROR = 11

-- data type
-- https://github.com/maxmind/libmaxminddb/blob/master/include/maxminddb.h#L40
local MMDB_DATA_TYPE_EXTENDED = 0
local MMDB_DATA_TYPE_POINTER = 1
local MMDB_DATA_TYPE_UTF8_STRING = 2
local MMDB_DATA_TYPE_DOUBLE = 3
local MMDB_DATA_TYPE_BYTES = 4
local MMDB_DATA_TYPE_UINT16 = 5
local MMDB_DATA_TYPE_UINT32 = 6
local MMDB_DATA_TYPE_MAP = 7
local MMDB_DATA_TYPE_INT32 = 8
local MMDB_DATA_TYPE_UINT64 = 9
local MMDB_DATA_TYPE_UINT128 = 10
local MMDB_DATA_TYPE_ARRAY = 11
local MMDB_DATA_TYPE_CONTAINER = 12
local MMDB_DATA_TYPE_END_MARKER = 13
local MMDB_DATA_TYPE_BOOLEAN = 14
local MMDB_DATA_TYPE_FLOAT = 15

-- copy from https://github.com/lilien1010/lua-resty-maxminddb/blob/f96633e2428f8f7bcc1e2a7a28b747b33233a8db/resty/maxminddb.lua#L136-L138

local initted = false

local function mmdb_strerror(profile, rc)
    return ffi_str(_D[profile].maxm.MMDB_strerror(rc))
end

local function gai_strerror(rc)
    return ffi_str(C.gai_strerror(rc))
end

function _M.init(profiles)
    for profile, location in pairs(profiles) do
        _D[profile] = {}
        _D[profile].maxm = ffi.load('/usr/local/openresty/1pwaf/data/libmaxminddb.so')
        _D[profile].mmdb = ffi_new('MMDB_s')
        local maxmind_ready = _D[profile].maxm.MMDB_open(location, 0, _D[profile].mmdb)
        if maxmind_ready ~= MMDB_SUCCESS then
            return nil, mmdb_strerror(profile, maxmind_ready)
        end
        ffi_gc(_D[profile].mmdb, _D[profile].maxm.MMDB_close)
    end

    --if not initted then
    --    local maxmind_ready = maxm.MMDB_open(dbfile, 0, mmdb)
    --
    --    if maxmind_ready ~= MMDB_SUCCESS then
    --        return nil, mmdb_strerror(maxmind_ready)
    --    end
    --
    --
    --
    --    ffi_gc(mmdb, maxm.MMDB_close)
    --end
    initted = true
    return initted
end

function _M.initted()
    return initted
end

-- https://github.com/maxmind/libmaxminddb/blob/master/src/maxminddb.c#L1938
-- LOCAL MMDB_entry_data_list_s *dump_entry_data_list( FILE *stream, MMDB_entry_data_list_s *entry_data_list, int indent, int *status)
local function _dump_entry_data_list(entry_data_list, status)

    if not entry_data_list then
        return nil, MMDB_INVALID_DATA_ERROR
    end

    local entry_data_item = entry_data_list[0].entry_data
    local data_type = entry_data_item.type
    local data_size = entry_data_item.data_size
    local result

    if data_type == MMDB_DATA_TYPE_MAP then
        result = {}

        local size = entry_data_item.data_size

        entry_data_list = entry_data_list[0].next

        while (size > 0 and entry_data_list)
        do
            entry_data_item = entry_data_list[0].entry_data
            data_type = entry_data_item.type
            data_size = entry_data_item.data_size

            if MMDB_DATA_TYPE_UTF8_STRING ~= data_type then
                return nil, MMDB_INVALID_DATA_ERROR
            end

            local key = ffi_str(entry_data_item.utf8_string, data_size)

            if not key then
                return nil, MMDB_OUT_OF_MEMORY_ERROR
            end

            local val
            entry_data_list = entry_data_list[0].next
            entry_data_list, status, val = _dump_entry_data_list(entry_data_list)

            if status ~= MMDB_SUCCESS then
                return nil, status
            end

            result[key] = val

            size = size - 1
        end


    elseif entry_data_list[0].entry_data.type == MMDB_DATA_TYPE_ARRAY then
        local size = entry_data_list[0].entry_data.data_size
        result = {}

        entry_data_list = entry_data_list[0].next

        local i = 1
        while (i <= size and entry_data_list)
        do
            local val
            entry_data_list, status, val = _dump_entry_data_list(entry_data_list)

            if status ~= MMDB_SUCCESS then
                return nil, nil, val
            end

            result[i] = val
            i = i + 1
        end


    else
        entry_data_item = entry_data_list[0].entry_data
        data_type = entry_data_item.type
        data_size = entry_data_item.data_size

        local val
        -- string type "key":"val"
        -- other type "key":val
        -- default other type
        if data_type == MMDB_DATA_TYPE_UTF8_STRING then
            val = ffi_str(entry_data_item.utf8_string, data_size)
            if not val then
                status = MMDB_OUT_OF_MEMORY_ERROR
                return nil, status
            end
        elseif data_type == MMDB_DATA_TYPE_BYTES then
            val = ffi_str(ffi_cast('char * ', entry_data_item.bytes), data_size)
            if not val then
                status = MMDB_OUT_OF_MEMORY_ERROR
                return nil, status
            end
        elseif data_type == MMDB_DATA_TYPE_DOUBLE then
            val = entry_data_item.double_value
        elseif data_type == MMDB_DATA_TYPE_FLOAT then
            val = entry_data_item.float_value
        elseif data_type == MMDB_DATA_TYPE_UINT16 then
            val = entry_data_item.uint16
        elseif data_type == MMDB_DATA_TYPE_UINT32 then
            val = entry_data_item.uint32
        elseif data_type == MMDB_DATA_TYPE_BOOLEAN then
            val = entry_data_item.boolean
        elseif data_type == MMDB_DATA_TYPE_UINT64 then
            val = entry_data_item.uint64
        elseif data_type == MMDB_DATA_TYPE_INT32 then
            val = entry_data_item.int32
        else
            return nil, MMDB_INVALID_DATA_ERROR
        end

        result = val
        entry_data_list = entry_data_list[0].next
    end

    status = MMDB_SUCCESS
    return entry_data_list, status, result
end

function _M.lookup(profile, ip)

    if not initted then
        return nil, "not initialized"
    end

    -- copy from https://github.com/lilien1010/lua-resty-maxminddb/blob/f96633e2428f8f7bcc1e2a7a28b747b33233a8db/resty/maxminddb.lua#L159-L176
    local gai_error = ffi_new('int[1]')
    local mmdb_error = ffi_new('int[1]')

    local result = _D[profile].maxm.MMDB_lookup_string(_D[profile].mmdb, ip, gai_error, mmdb_error)

    if mmdb_error[0] ~= MMDB_SUCCESS then
        return nil, 'lookup failed: ' .. mmdb_strerror(profile, mmdb_error[0])
    end

    if gai_error[0] ~= MMDB_SUCCESS then
        return nil, 'lookup failed: ' .. gai_strerror(gai_error[0])
    end

    if true ~= result.found_entry then
        return nil, 'not found'
    end

    local entry_data_list = ffi_cast('MMDB_entry_data_list_s **const', ffi_new("MMDB_entry_data_list_s"))

    local status = _D[profile].maxm.MMDB_get_entry_data_list(result.entry, entry_data_list)

    if status ~= MMDB_SUCCESS then
        return nil, 'get entry data failed: ' .. mmdb_strerror(profile, status)
    end

    local head = entry_data_list[0] -- Save so this can be passed to free fn.
    local _, status, result = _dump_entry_data_list(entry_data_list)
    _D[profile].maxm.MMDB_free_entry_data_list(head)

    if status ~= MMDB_SUCCESS then
        return nil, 'dump entry data failed: ' .. mmdb_strerror(profile, status)
    end

    return result
end

-- copy from https://github.com/lilien1010/lua-resty-maxminddb/blob/master/resty/maxminddb.lua#L208
-- https://www.maxmind.com/en/geoip2-databases  you should download  the mmdb file from maxmind

return _M;
