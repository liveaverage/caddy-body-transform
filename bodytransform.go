package bodytransform

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "net/http"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    lua "github.com/yuin/gopher-lua"
    json "github.com/layeh/gopher-json"
)

func init() {
    caddy.RegisterModule(BodyTransform{})
}

// BodyTransform is a Caddy module that transforms the request or response body using a Lua script
type BodyTransform struct {
    Script       string `json:"script,omitempty"`
    TransformType string `json:"transform_type,omitempty"` // "request" or "response"
    luaState     *lua.LState
}

// CaddyModule returns the Caddy module information.
func (BodyTransform) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.body_transform",
        New: func() caddy.Module { return new(BodyTransform) },
    }
}

// Provision sets up the module. Loads the Lua script.
func (bt *BodyTransform) Provision(ctx caddy.Context) error {
    bt.luaState = lua.NewState()
    json.Preload(bt.luaState)
    if err := bt.luaState.DoString(bt.Script); err != nil {
        return fmt.Errorf("failed to load Lua script: %v", err)
    }
    return nil
}

// ServeHTTP implements the caddyhttp.MiddlewareHandler interface.
func (bt *BodyTransform) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    // Initial log to confirm the handler is called
    fmt.Println("BodyTransform ServeHTTP called")

    if bt.TransformType == "response" {
        // Capture the response
        recorder := &responseRecorder{ResponseWriter: w, body: &bytes.Buffer{}}
        err := next.ServeHTTP(recorder, r)
        if err != nil {
            return err
        }

        // Log that the response transformation is being executed
        fmt.Println("Executing response transformation")

        // Log the captured response body
        fmt.Println("Captured response body: " + recorder.body.String())

        // Transform the response body
        transformedBody, err := bt.transform(recorder.body.Bytes())
        if err != nil {
            return err
        }

        // Write the transformed response
        for k, v := range recorder.Header() {
            w.Header()[k] = v
        }
        w.Header().Set("Content-Length", fmt.Sprint(len(transformedBody)))
        w.WriteHeader(recorder.statusCode)
        _, err = w.Write(transformedBody)
        if err != nil {
            return err
        }

        // Log the transformed response body
        fmt.Println("Transformed response body: " + string(transformedBody))
        return nil
    }

    // Transform the request body
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }
    transformedBody, err := bt.transform(body)
    if err != nil {
        return err
    }

    r.Body = ioutil.NopCloser(bytes.NewReader(transformedBody))
    r.ContentLength = int64(len(transformedBody))

    // Proceed with the next handler
    return next.ServeHTTP(w, r)
}

// transform executes the Lua script to transform the body
func (bt *BodyTransform) transform(body []byte) ([]byte, error) {
    // Create a new Lua state for this transformation
    L := lua.NewState()
    defer L.Close()
    json.Preload(L)

    // Load the script into the new Lua state
    if err := L.DoString(bt.Script); err != nil {
        return nil, fmt.Errorf("failed to load Lua script: %v", err)
    }

    // Push the body onto the Lua stack
    L.Push(lua.LString(string(body)))

    // Call the Lua function
    if err := L.CallByParam(lua.P{
        Fn:      L.GetGlobal("transform"),
        NRet:    1,
        Protect: true,
    }, L.Get(-1)); err != nil {
        return nil, fmt.Errorf("failed to call Lua function: %v", err)
    }

    // Get the transformed body
    transformedBody := L.Get(-1).(lua.LString)
    L.Pop(1)

    return []byte(transformedBody), nil
}

// responseRecorder is a wrapper for http.ResponseWriter to capture the response
type responseRecorder struct {
    http.ResponseWriter
    statusCode int
    body       *bytes.Buffer
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
    rec.statusCode = statusCode
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
    rec.body.Write(b)
    return rec.ResponseWriter.Write(b)
}

// Interface guard
var (
    _ caddyhttp.MiddlewareHandler = (*BodyTransform)(nil)
)
