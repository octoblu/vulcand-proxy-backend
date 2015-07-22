package backendheader

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mailgun/vulcand/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/mailgun/vulcand/plugin"
)

const Type = "backendheader"

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      Type,       // A short name for the middleware
		FromOther: FromOther,  // Tells vulcand how to rcreate middleware from another one (this is for deserialization)
		FromCli:   FromCli,    // Tells vulcand how to create middleware from command line tool
		CliFlags:  CliFlags(), // Vulcand will add this flags to middleware specific command line tool
	}
}

type BackendHeaderMiddleware struct {
	AddHeader bool
	HeaderName string
}

// Auth middleware handler
type BackendHeaderHandler struct {
	next http.Handler
	AddHeader bool
	HeaderName string
}

// This function will be called each time the request hits the location with this middleware activated
func (h *BackendHeaderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.AddHeader {
		w.Header().Set(h.HeaderName, GetLocalIP())
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
func New(addHeader bool, headerName string) (*BackendHeaderMiddleware, error) {
	return &BackendHeaderMiddleware{AddHeader: addHeader, HeaderName: headerName}, nil
}

// This function is important, it's called by vulcand to create a new handler from the middleware config and put it into the
// middleware chain. Note that we need to remember 'next' handler to call
func (c *BackendHeaderMiddleware) NewHandler(next http.Handler) (http.Handler, error) {
	return &BackendHeaderHandler{next: next, AddHeader: c.AddHeader, HeaderName: c.HeaderName}, nil
}

// String() will be called by loggers inside Vulcand and command line tool.
func (c *BackendHeaderMiddleware) String() string {
	return fmt.Sprintf("%v, addHeader=%v", c.HeaderName, c.AddHeader)
}

// FromOther Will be called by Vulcand when engine or API will read the middleware from the serialized format.
// It's important that the signature of the function will be exactly the same, otherwise Vulcand will
// fail to register this middleware.
// The first and the only parameter should be the struct itself, no pointers and other variables.
// Function should return middleware interface and error in case if the parameters are wrong.
func FromOther(c BackendHeaderMiddleware) (plugin.Middleware, error) {
	return New(c.AddHeader, c.HeaderName)
}

// FromCli constructs the middleware from the command line
func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return New(c.Bool("addHeader"), c.String("headerName"))
}

// CliFlags will be used by Vulcand construct help and CLI command for the vctl command
func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "addHeader",
			Usage: "if provided, add X-Backend-Server to the response",
		},
		cli.StringFlag{
			Name:  "headerName",
			Value: "X-Backend-Server",
			Usage: "defaults to X-Backend-Server",
		},
	}
}
