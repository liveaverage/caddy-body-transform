# caddy-body-transform

## Usage


### Build

Build with:
```
xcaddy build --with github.com/liveaverage/caddy-body-transform
```

### Run

```
          "routes": [
            {
              "match": [{"path": ["/predict*"]}],
              "handle": [
                {
                  "handler": "body_transform",
                  "script": "function transform(body) local json = require 'cjson' local data = json.decode(body) local first_instance = data.instances[1] return json.encode(first_instance) end"
                },
                {
                  "handler": "reverse_proxy",
                  "upstreams": [{"dial": "127.0.0.1:${BACKEND_PORT}"}]
                }
              ]
```