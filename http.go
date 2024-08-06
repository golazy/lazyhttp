package lazyhttp

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	"golazy.dev/lazyapp"
)

type HttpService struct {
	http.Server
}

type errSlog2Log struct {
	*slog.Logger
}

func (l *errSlog2Log) Write(p []byte) (n int, err error) {
	l.Error(string(p))
	return len(p), nil
}

func (s *HttpService) Run(ctx context.Context, l *slog.Logger) error {
	s.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}
	s.ReadHeaderTimeout = time.Millisecond * 200
	s.ErrorLog = log.New(&errSlog2Log{l}, "", 0)
	s.DisableGeneralOptionsHandler = true

	errCh := make(chan error)
	go func() {
		<-ctx.Done()
		sctx, _ := context.WithTimeout(context.Background(), time.Second*5)

		errCh <- s.Shutdown(sctx)
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return <-errCh

}

type serviceDesc struct {
	name string
}

func (d serviceDesc) Name() string {
	return d.name
}

func (s *HttpService) Desc() lazyapp.ServiceDescription {
	return serviceDesc{name: "lazyhttp"}
}
