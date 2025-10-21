package httpserver

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/timmbarton/utils/types/secs"
	"go.uber.org/zap"
)

type Config struct {
	StartTimeout secs.Seconds `validate:"seconds"`
	StopTimeout  secs.Seconds `validate:"seconds"`

	Addr                        string `validate:"required"`
	ServiceId                   int    `validate:"required,min=10,max=99"`
	ShowUnknownErrorsInResponse bool
	FiberConfig                 *fiber.Config
	Logger                      *zap.Logger
}

type DefaultServer struct {
	cfg   Config
	fiber *fiber.App
}

var ErrStopTimeOut = errors.New("stop timeout")

func (s *DefaultServer) Init(cfg Config, bind func(fiber.Router)) {
	fiberCfg := fiber.Config{DisableStartupMessage: true}
	if cfg.FiberConfig != nil {
		fiberCfg = *cfg.FiberConfig
	}

	s.fiber = fiber.New(fiberCfg)
	s.cfg = cfg
	s.fiber.Use(
		logger.New(
			logger.Config{
				Format:     "${pid} ${status} - ${method} ${path}\n",
				TimeFormat: time.DateTime,
				TimeZone:   "Europe/Moscow",
			},
		),
	)
	s.fiber.Use(GetErrsMiddleware(cfg.ServiceId, cfg.ShowUnknownErrorsInResponse, cfg.Logger))

	bind(s.fiber)
}

func (s *DefaultServer) Start(_ context.Context) error {
	errCh := make(chan error)

	go func() {
		if err := s.fiber.Listen(s.cfg.Addr); err != nil {
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
	okCh, errCh := make(chan any), make(chan error)

	go func() {
		if err := s.fiber.Shutdown(); err != nil {
			errCh <- err
		}

		okCh <- nil
	}()

	select {
	case <-okCh:
		return nil
	case err := <-errCh:
		return err
	case <-time.After(time.Duration(s.cfg.StopTimeout)):
		return ErrStopTimeOut
	}
}

func (s *DefaultServer) GetName() string { return "HTTP Server" }
