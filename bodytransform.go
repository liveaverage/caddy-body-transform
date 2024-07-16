package bodytransform

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    lua "github.com/yuin/gopher-lua"
)

func init() {
    caddy.RegisterModule(BodyTransform{})
}

// BodyTransform is a Caddy module that transforms the request body using a Lua script
type BodyTransform struct {
    Script string `json:"script,omitempty"`
    luaState *lua.LState
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
    if err := bt.luaState.DoString(bt.Script); err != nil {
        return fmt.Errorf("failed to load Lua script: %v", err)
    }
    return nil
}

// ServeHTTP implements the caddyhttp.MiddlewareHandler interface.
func (bt BodyTransform) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    // Read the body
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        return err
    }

    // Create a new Lua state for this request
    L := bt.luaState.NewThread()

    // Push the body onto the Lua stack
    L.Push(lua.LString(string(body)))

    // Call the Lua function
    if err := L.CallByParam(lua.P{
        Fn:      L.GetGlobal("transform"),
        NRet:    1,
        Protect: true,
    }, L.Get(-1)); err != nil {
        return fmt.Errorf("failed to call Lua function: %v", err)
    }

    // Get the transformed body
    transformedBody := L.Get(-1).(lua.LString)
    L.Pop(1)

    // Update the request body
    r.Body = ioutil.NopCloser(bytes.NewReader([]byte(transformedBody)))
    r.ContentLength = int64(len(transformedBody))

    // Proceed with the next handler
    return next.ServeHTTP(w, r)
}

// Interface guard
var (
    _ caddyhttp.MiddlewareHandler = (*BodyTransform)(nil)
)
