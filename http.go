package http_default

import (
	"context"
	"fmt"
	nethttp "net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	. "github.com/infrago/base"
	"github.com/infrago/http"
)

//------------------------- 默认事件驱动 begin --------------------------

const (
	defaultSeparator = "|||"
)

type (
	defaultDriver  struct{}
	defaultConnect struct {
		mutex   sync.RWMutex
		actives int64

		config http.Config

		server *nethttp.Server
		router *mux.Router

		delegate http.Delegate

		routes map[string]*mux.Route
	}

	//响应对象
	defaultThread struct {
		connect  *defaultConnect
		name     string
		site     string
		params   Map
		request  *nethttp.Request
		response nethttp.ResponseWriter
	}
)

// 连接
func (driver *defaultDriver) Connect(config http.Config) (http.Connect, error) {
	return &defaultConnect{
		config: config, routes: map[string]*mux.Route{},
	}, nil
}

// 打开连接
func (this *defaultConnect) Open() error {
	this.router = mux.NewRouter()
	this.server = &nethttp.Server{
		Addr:         fmt.Sprintf("%s:%d", this.config.Host, this.config.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      this.router,
	}

	return nil
}
func (this *defaultConnect) Health() (http.Health, error) {
	//this.mutex.RLock()
	//defer this.mutex.RUnlock()
	return http.Health{Workload: this.actives}, nil
}

// 关闭连接
func (this *defaultConnect) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return this.server.Shutdown(ctx)
}

// 注册回调
func (this *defaultConnect) Accept(delegate http.Delegate) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	//保存回调
	this.delegate = delegate

	//先注册一个接入全部请求的
	this.router.NotFoundHandler = this
	this.router.MethodNotAllowedHandler = this

	return nil
}

// 订阅者，注册事件
func (this *defaultConnect) Register(name string, info http.Info) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	route := this.router.HandleFunc(info.Uri, this.ServeHTTP).Name(name)
	for _, host := range info.Hosts {
		route.Host(host)
	}
	if info.Method != "" {
		route.Methods(info.Method)
	}

	this.routes[name] = route

	return nil
}

func (this *defaultConnect) Start() error {
	if this.server == nil {
		panic("Invalid http this.")
	}

	go func() {
		err := this.server.ListenAndServe()
		if err != nil && err != nethttp.ErrServerClosed {
			panic(err.Error())
		}
	}()

	return nil
}
func (this *defaultConnect) StartTLS(certFile, keyFile string) error {
	if this.server == nil {
		panic("Invalid http this.")
	}

	go func() {
		err := this.server.ListenAndServeTLS(certFile, keyFile)
		if err != nil && err != nethttp.ErrServerClosed {
			panic(err.Error())
		}
	}()

	return nil
}

func (this *defaultConnect) ServeHTTP(res nethttp.ResponseWriter, req *nethttp.Request) {
	name := ""
	params := Map{}

	route := mux.CurrentRoute(req)

	if route != nil {
		name = route.GetName()
		vars := mux.Vars(req)
		for k, v := range vars {
			params[k] = v
		}
	}

	// 有请求都发，404也转过去
	this.delegate.Serve(name, params, res, req)
}

//------------------------- 默认HTTP驱动 end --------------------------
