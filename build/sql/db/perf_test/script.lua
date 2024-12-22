local json = require("cjson")
local data = {}
init = function(args)
    counter = 0
    local filename = "data.json"
    local file = io.open(filename, "r")
    local content = file:read("*a")
    data = json.decode(content)
    file:close()
end

request = function()
    counter = counter + 1
    local item = data[counter%#data]
    return wrk.format("POST", "/api/auth/signup", nil, json.encode(item))
end
