package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"

	"github.com/timmbarton/errors"
	"github.com/timmbarton/utils/types/secs"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type DefaultServerConfig struct {
	StartTimeout secs.Seconds `validate:"seconds"`
	StopTimeout  secs.Seconds `validate:"seconds"`

	Host              string       `validate:"required,min=1"`
	ServiceId         int          `validate:"required,min=10,max=99"`
	MaxConnectionIdle secs.Seconds `validate:"seconds"`
	Timeout           secs.Seconds `validate:"seconds"`
	MaxConnectionAge  secs.Seconds `validate:"seconds"`
	Time              secs.Seconds `validate:"seconds"`

	DisableReflection bool
}

type DefaultServer struct {
	cfg        DefaultServerConfig
	grpcServer *grpc.Server
	listener   net.Listener
}

type logger interface {
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

func (s *DefaultServer) Init(cfg DefaultServerConfig, logger logger) {
	s.cfg = cfg

	s.grpcServer = grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Duration(s.cfg.MaxConnectionIdle),
			Timeout:           time.Duration(s.cfg.Timeout),
			MaxConnectionAge:  time.Duration(s.cfg.MaxConnectionAge),
			Time:              time.Duration(s.cfg.Time),
		}),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			errs.GetGRPCInterceptor(s.cfg.ServiceId),
			errs.GetLoggingInterceptor(logger),
		),
	)
}

func (s *DefaultServer) RegisterService(sd *grpc.ServiceDesc, ss any) {
	s.grpcServer.RegisterService(sd, ss)
}

func (s *DefaultServer) Start(_ context.Context) error {
	if !s.cfg.DisableReflection {
		reflection.Register(s.grpcServer)
	}

	errCh := make(chan error)

	go func() {
		err := error(nil)

		s.listener, err = net.Listen("tcp", s.cfg.Host)
		if err != nil {
			errCh <- err
			return
		}

		err = s.grpcServer.Serve(s.listener)
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(time.Duration(s.cfg.StartTimeout)):
		return nil
	}
}
func (s *DefaultServer) Stop(_ context.Context) error {
	stopCh := make(chan any)
	go func() {
		s.grpcServer.GracefulStop()
		stopCh <- nil
	}()

	select {
	case <-time.After(time.Duration(s.cfg.StopTimeout)):
		return nil
	case <-stopCh:
		return nil
	}
}
func (s *DefaultServer) GetName() string { return fmt.Sprintf("GRPC Server at %s", s.cfg.Host) }
