local match = string.match
local ngxmatch=ngx.re.match
local unescape=ngx.unescape_uri
local get_headers = ngx.req.get_headers
local cjson = require "cjson"
local content_length=tonumber(ngx.req.get_headers()['content-length'])
local method=ngx.req.get_method()


local function optionIsOn(options)
	return options == "on" or options == "On" or options == "ON"
end

local logpath = ngx.var.logdir
local rulepath = ngx.var.RulePath
local attacklog = optionIsOn(ngx.var.attackLog)
local Redirect=optionIsOn(ngx.var.redirect)
local CCDeny = optionIsOn(ngx.var.CCDeny)
local UrlBlockDeny = optionIsOn(ngx.var.urlBlockDeny)
local UrlWhiteAllow = optionIsOn(ngx.var.urlWhiteAllow)
local IpBlockDeny = optionIsOn(ngx.var.ipBlockDeny)
local IpWhiteAllow = optionIsOn(ngx.var.ipWhiteAllow)
local PostDeny = optionIsOn(ngx.var.postDeny)
local ArgsDeny = optionIsOn(ngx.var.argsDeny)
local CookieDeny = optionIsOn(ngx.var.cookieDeny)
local FileExtDeny = optionIsOn(ngx.var.fileExtDeny)

local function getClientIp()
	IP  = ngx.var.remote_addr
	if IP == nil then
		IP  = "unknown"
	end
	return IP
end
local function write(logfile,msg)
	local fd = io.open(logfile,"ab")
	if fd == nil then return end
	fd:write(msg)
	fd:flush()
	fd:close()
end
local function log(method,url,data,ruletag)
	if attacklog then
		local realIp = getClientIp()
		local ua = ngx.var.http_user_agent
		local servername=ngx.var.server_name
		local time=ngx.localtime()
        local line = nil
		if ua  then
			line = realIp.." ["..time.."] \""..method.." "..servername..url.."\" \""..data.."\"  \""..ua.."\" \""..ruletag.."\"\n"
		else
			line = realIp.." ["..time.."] \""..method.." "..servername..url.."\" \""..data.."\" - \""..ruletag.."\"\n"
		end
		local filename = logpath..'/'..servername.."_"..ngx.today().."_sec.log"
		write(filename,line)
	end
end
------------------------------------规则读取函数-------------------------------------------------------------------
local function read_rule(var)
	file = io.open(rulepath..'/'..var,"r")
	if file==nil then
		return
	end
	t = {}
	for line in file:lines() do
		table.insert(t,line)
	end
	file:close()
	return(t)
end

local function read_json(var)
    file = io.open(rulepath..'/'..var,"r")
	if file==nil then
		return
	end
    str = file:read("*a")
    file:close()
    list = cjson.decode(str)
    return list
end

local function read_str(var)
    file = io.open(rulepath..'/'..var,"r")
	if file==nil then
		return
	end
    local str = file:read("*a")
    file:close()
	return str
end



local urlWhiteList=read_rule('urlWhiteList')
local urlBlockList=read_rule('urlBlockList')
local argsCheckList=read_rule('argsCheckList')
local postCheckList=read_rule('postCheckList')
local cookieBlockList=read_rule('cookieBlockList')
local ipWhiteList=read_json('ipWhiteList')
local ipBlockList=read_json('ipBlockList')
local ccRate=read_str('ccRate')
local fileExtBlockList = read_json('fileExtBlockList')

local html=read_str('html')
local uarules=read_rule('user-agent')

local function say_html()
	if Redirect then
		ngx.header.content_type = "text/html"
		ngx.status = ngx.HTTP_FORBIDDEN
		ngx.say(html)
		ngx.exit(ngx.status)
	end
end

local function whiteurl()
	if UrlWhiteAllow then
		if urlWhiteList ~=nil then
			for _,rule in pairs(urlWhiteList) do
				if ngxmatch(ngx.var.uri,rule,"isjo") then
					return true
				end
			end
		end
	end
	return false
end
local function fileExtCheck(ext)
	if FileExtDeny then
		local items = Set(fileExtBlockList)
		ext=string.lower(ext)
		if ext then
			for rule in pairs(items) do
				if ngx.re.match(ext,rule,"isjo") then
					log('POST',ngx.var.request_uri,"-","file attack with ext "..ext)
					say_html()
				end
			end
		end
	end
	return false
end
function Set (list)
	local set = {}
	for _, l in ipairs(list) do set[l] = true end
	return set
end

local function args()
	if ArgsDeny then
		if argsCheckList then
			for _,rule in pairs(argsCheckList) do
				local uriArgs = ngx.req.get_uri_args()
				for key, val in pairs(uriArgs) do
					if type(val)=='table' then
						local t={}
						for k,v in pairs(val) do
							if v == true then
								v=""
							end
							table.insert(t,v)
						end
						data=table.concat(t, " ")
					else
						data=val
					end
					if data and type(data) ~= "boolean" and rule ~="" and ngxmatch(unescape(data),rule,"isjo") then
						log('GET',ngx.var.request_uri,"-",rule)
						say_html()
						return true
					end
				end
			end
		end
	end
	return false
end


local function url()
	if UrlBlockDeny then
		for _,rule in pairs(urlBlockList) do
			if rule ~="" and ngxmatch(ngx.var.request_uri,rule,"isjo") then
				log('GET',ngx.var.request_uri,"-",rule)
				say_html()
				return true
			end
		end
	end
	return false
end

function ua()
	local ua = ngx.var.http_user_agent
	if ua ~= nil then
		for _,rule in pairs(uarules) do
			if rule ~="" and ngxmatch(ua,rule,"isjo") then
				log('UA',ngx.var.request_uri,"-",rule)
				say_html()
				return true
			end
		end
	end
	return false
end
function body(data)
	for _,rule in pairs(postCheckList) do
		if rule ~="" and data~="" and ngxmatch(unescape(data),rule,"isjo") then
			log('POST',ngx.var.request_uri,data,rule)
			say_html()
			return true
		end
	end
	return false
end
local function cookie()
	local ck = ngx.var.http_cookie
	if CookieDeny and ck then
		for _,rule in pairs(cookieBlockList) do
			if rule ~="" and ngxmatch(ck,rule,"isjo") then
				log('Cookie',ngx.var.request_uri,"-",rule)
				say_html()
				return true
			end
		end
	end
	return false
end

local function denycc()
	if CCDeny and ccRate then
		local uri=ngx.var.uri
		CCcount=tonumber(string.match(ccRate,'(.*)/'))
		CCseconds=tonumber(string.match(ccRate,'/(.*)'))
		local uri = getClientIp()..uri
		local limit = ngx.shared.limit
		local req,_=limit:get(uri)
		if req then
			if req > CCcount then
				ngx.exit(503)
				return true
			else
				limit:incr(token,1)
			end
		else
			limit:set(uri,1,CCseconds)
		end
	end
	return false
end

local function get_boundary()
	local header = get_headers()["content-type"]
	if not header then
		return nil
	end

	if type(header) == "table" then
		header = header[1]
	end

	local m = match(header, ";%s*boundary=\"([^\"]+)\"")
	if m then
		return m
	end

	return match(header, ";%s*boundary=([^\",;]+)")
end

local function whiteip()
	if IpWhiteAllow then
		if next(ipWhiteList) ~= nil then
			for _,ip in pairs(ipWhiteList) do
				if getClientIp()==ip then
					return true
				end
			end
		end
	end
	return false
end

local function blockip()
	if IpBlockDeny then
		if next(ipBlockList) ~= nil then
			for _,ip in pairs(ipBlockList) do
				if getClientIp()==ip then
					ngx.exit(403)
					return true
				end
			end
		end
	end
	return false
end



if whiteip() then
elseif blockip() then
elseif denycc() then
elseif ngx.var.http_Acunetix_Aspect then
    ngx.exit(444)
elseif ngx.var.http_X_Scan_Memo then
    ngx.exit(444)
elseif whiteurl() then
elseif ua() then
elseif url() then
elseif args() then
elseif cookie() then
elseif PostDeny then
    if method=="POST" then   
            local boundary = get_boundary()
	    if boundary then
	    local len = string.len
            local sock, err = ngx.req.socket()
    	    if not sock then
					return
            end
	    ngx.req.init_body(128 * 1024)
            sock:settimeout(0)
	    local content_length = nil
    	    content_length=tonumber(ngx.req.get_headers()['content-length'])
    	    local chunk_size = 4096
            if content_length < chunk_size then
					chunk_size = content_length
	    end
            local size = 0
	    while size < content_length do
		local data, err, partial = sock:receive(chunk_size)
		data = data or partial
		if not data then
			return
		end
		ngx.req.append_body(data)
        	if body(data) then
	   	        return true
    	    	end
		size = size + len(data)
		local m = ngxmatch(data,[[Content-Disposition: form-data;(.+)filename="(.+)\\.(.*)"]],'ijo')
        	if m then
            		fileExtCheck(m[3])
            		filetranslate = true
        	else
            		if ngxmatch(data,"Content-Disposition:",'isjo') then
                		filetranslate = false
            		end
            		if filetranslate==false then
            			if body(data) then
                    			return true
                		end
            		end
        	end
		local less = content_length - size
		if less < chunk_size then
			chunk_size = less
		end
	 end
	 ngx.req.finish_body()
    else
			ngx.req.read_body()
			local args = ngx.req.get_post_args()
			if not args then
				return
			end
			for key, val in pairs(args) do
				if type(val) == "table" then
					if type(val[1]) == "boolean" then
						return
					end
					data=table.concat(val, ", ")
				else
					data=val
				end
				if data and type(data) ~= "boolean" and body(data) then
                			body(key)
				end
			end
		end
    end
else
    return
end
