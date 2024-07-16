# caddy-body-transform

## Usage


### Build

Build with:
```
xcaddy build --with github.com/liveaverage/caddy-body-transform
```

### Run

Sample configuration snippet, extracting the first element of list `instances[]` and passing to defined upstream server: 
```
              "match": [{"path": ["/predict*"]}],
              "handle": [
                {
                  "handler": "body_transform",
                  "script": "function transform(body) local json = require 'json' local data = json.decode(body) local first_instance = data.instances[1] return json.encode(first_instance) end"
                },
                {
                  "handler": "static_response",
                  "body": "{http.request.body}",
                  "headers": {
                    "Content-Type": ["application/json"]
                  }
                }
              ]
```