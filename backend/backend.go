package backend

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mailgun/vulcand/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/mailgun/vulcand/plugin"
)

const Type = "backend"

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      Type,       // A short name for the middleware
		FromOther: FromOther,  // Tells vulcand how to rcreate middleware from another one (this is for deserialization)
		FromCli:   FromCli,    // Tells vulcand how to create middleware from command line tool
		CliFlags:  CliFlags(), // Vulcand will add this flags to middleware specific command line tool
	}
}

type BackendMiddleware struct {
	addHeader bool
}

// Auth middleware handler
type BackendHandler struct {
	cfg BackendMiddleware
	next http.Handler
}

// This function will be called each time the request hits the location with this middleware activated
func (h *BackendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.cfg.addHeader {
		w.Header().Set("X-Backend-Server", GetLocalIP())
	}
	h.next.ServeHTTP(w, r)
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

// This function is optional but handy, used to check input parameters when creating new middlewares
func New(addHeader bool) (*BackendMiddleware, error) {
	return &BackendMiddleware{addHeader: addHeader}, nil
}

// This function is important, it's called by vulcand to create a new handler from the middleware config and put it into the
// middleware chain. Note that we need to remember 'next' handler to call
func (c *BackendMiddleware) NewHandler(next http.Handler) (http.Handler, error) {
	return &BackendHandler{next: next, cfg: *c}, nil
}

// String() will be called by loggers inside Vulcand and command line tool.
func (c *BackendMiddleware) String() string {
	return fmt.Sprintf("Adding X-Backend-Server")
}

// FromOther Will be called by Vulcand when engine or API will read the middleware from the serialized format.
// It's important that the signature of the function will be exactly the same, otherwise Vulcand will
// fail to register this middleware.
// The first and the only parameter should be the struct itself, no pointers and other variables.
// Function should return middleware interface and error in case if the parameters are wrong.
func FromOther(c BackendMiddleware) (plugin.Middleware, error) {
	return New(c.addHeader)
}

// FromCli constructs the middleware from the command line
func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return New(c.Bool("add-backend-header"))
}

// CliFlags will be used by Vulcand construct help and CLI command for the vctl command
func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "add-backend-header",
			Usage: "if provided, add X-Backend-Server to the response",
		},
	}
}
