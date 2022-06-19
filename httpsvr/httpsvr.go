package httpsvr

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/OutOfBedlam/ots/logging"
	"github.com/gin-gonic/gin"
)

type HttpServerConfig struct {
	DisableConsoleColor bool
	DebugMode           bool
	BindHostPort        string
	LoggingConfig       *logging.Config
	LoggerName          string
}

type HttpServer struct {
	httpServer *http.Server
	log        logging.Log

	ListenerHost string
	ListenerPort string
}

func NewServer(conf *HttpServerConfig) *HttpServer {
	if conf.DisableConsoleColor {
		gin.DisableConsoleColor()
	}
	if conf.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	httpLogger := logging.New(conf.LoggingConfig)

	ginLogConfig := gin.LoggerConfig{
		Formatter: logFormat,
		Output:    httpLogger,
		SkipPaths: []string{},
	}

	r := gin.New()
	r.Use(gin.LoggerWithConfig(ginLogConfig))
	r.Use(gin.Recovery())

	loggerName := conf.LoggerName
	if loggerName == "" {
		loggerName = "http"
	}

	svr := HttpServer{
		httpServer: &http.Server{
			Addr:    conf.BindHostPort,
			Handler: r,
		},
		log: logging.GetLog(loggerName),
	}

	return &svr
}

func (svr *HttpServer) Handler() http.Handler {
	return svr.httpServer.Handler
}

func (svr *HttpServer) GET(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	r := svr.httpServer.Handler.(*gin.Engine)
	return r.GET(path, handlers...)
}

func (svr *HttpServer) POST(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	r := svr.httpServer.Handler.(*gin.Engine)
	return r.POST(path, handlers...)
}

func (svr *HttpServer) PUT(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	r := svr.httpServer.Handler.(*gin.Engine)
	return r.PUT(path, handlers...)
}

func (svr *HttpServer) DELETE(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	r := svr.httpServer.Handler.(*gin.Engine)
	return r.DELETE(path, handlers...)
}

func (svr *HttpServer) Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	r := svr.httpServer.Handler.(*gin.Engine)
	return r.Any(relativePath, handlers...)
}

func (svr *HttpServer) Use(handler gin.HandlerFunc) gin.IRoutes {
	r := svr.httpServer.Handler.(*gin.Engine)
	return r.Use(handler)
}

func (svr *HttpServer) Start(opt ...interface{}) {
	var listener net.Listener
	var err error
	for _, o := range opt {
		switch v := o.(type) {
		case net.Listener:
			listener = v
		}
	}

	if listener == nil {
		listener, err = net.Listen("tcp", svr.httpServer.Addr)
		if err != nil {
			svr.log.Errorf("Server binding failed, %v", err)
			return
		}
	}

	if listener.Addr().Network() == "tcp" {
		svr.ListenerHost, svr.ListenerPort, _ = net.SplitHostPort(listener.Addr().String())
		svr.log.Infof("Listening on %v:%v", svr.ListenerHost, svr.ListenerPort)
	} else {
		svr.log.Infof("Listening on %v", listener.Addr())
	}
	go http.Serve(listener, svr.httpServer.Handler)
}

func (svr *HttpServer) Stop() {
	svr.log.Infof("Closing http server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svr.httpServer.Shutdown(ctx); err != nil {
		svr.log.Warnf("Server forced to shutdown: %s", err)
	}

	svr.log.Infof("Closed http server")
}
