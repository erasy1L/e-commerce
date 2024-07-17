package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type Server struct {
	http *http.Server

	grpc *grpc.Server

	listener net.Listener
}

type Configuration func(r *Server) error

func New(cfg ...Configuration) (r *Server, err error) {
	r = &Server{}

	for _, cfg := range cfg {
		if err = cfg(r); err != nil {
			return
		}
	}
	return
}

func (r *Server) Start() error {
	if r.http != nil {
		go func() {
			if err := r.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

	}

	if r.grpc != nil {
		go func() {
			if err := r.grpc.Serve(r.listener); err != nil {
				panic(err)
			}
		}()
	}

	return nil
}

func (r *Server) Stop(ctx context.Context) error {
	return r.http.Shutdown(ctx)
}

func WithHTTPServer(handler http.Handler, port string) Configuration {
	return func(r *Server) error {
		r.http = &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		}
		return nil
	}
}

func WithGRPCServer(server *grpc.Server, port string) Configuration {
	return func(r *Server) (err error) {
		r.listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
		r.grpc = server
		return err
	}
}
