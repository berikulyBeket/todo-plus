package httpserver

import (
	"net"
	"time"
)

// Option defines a type for functional options that configure the Server
type Option func(*Server)

// HostPort sets the server's address using the provided host and port
func HostPort(host string, port string) Option {
	return func(s *Server) {
		s.server.Addr = net.JoinHostPort(host, port)
	}
}

// ReadTimeout sets the server's maximum duration for reading the entire request, including the body
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

// WriteTimeout sets the server's maximum duration before timing out writes of the response
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}

// ShutdownTimeout sets the timeout for gracefully shutting down the server
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}
