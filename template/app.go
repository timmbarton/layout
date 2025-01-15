package template

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/timmbarton/layout/lifecycle"
)

var (
	ErrStartTimeout    = errors.New("start timeout")
	ErrShutdownTimeout = errors.New("shutdown timeout")
)

const (
	DefaultStartTimeout = 30 * time.Second
	DefaultStopTimeout  = 30 * time.Second
)

type App struct {
	components []lifecycle.Lifecycle

	startTimeout time.Duration
	stopTimeout  time.Duration
}

func (a *App) AddComponents(components ...lifecycle.Lifecycle) {
	a.components = append(a.components, components...)
}

func (a *App) Start(ctx context.Context) error {
	log.Println("starting app")

	okCh, errCh := make(chan any), make(chan error)

	// start app
	go func() {
		err := error(nil)
		// start each component
		for _, c := range a.components {
			log.Printf("starting %s...\n", c.GetName())

			err = c.Start(ctx)
			if err != nil {
				log.Printf("error on starting %s\n", c.GetName())
				errCh <- err

				return
			}
		}
		okCh <- nil
	}()

	select {
	case <-ctx.Done():
		return ErrStartTimeout
	case err := <-errCh:
		return err
	case <-okCh:
		log.Println("Application started!")
		return nil
	}
}
func (a *App) Stop(ctx context.Context) error {
	log.Println("shutting down service...")
	okCh, errCh := make(chan any), make(chan error)

	go func() {
		for i := len(a.components) - 1; i >= 0; i-- {
			c := a.components[i]
			log.Printf("stopping %s...\n", c.GetName())

			err := c.Stop(ctx)
			if err != nil {
				log.Println(err.Error())
				errCh <- err

				return
			}
		}
		okCh <- nil
	}()

	select {
	case <-ctx.Done():
		return ErrShutdownTimeout
	case err := <-errCh:
		return err
	case <-okCh:
		log.Println("Application stopped!")
		return nil
	}
}

func (a *App) GetStartTimeout() time.Duration {
	if a.startTimeout > 0 {
		return a.startTimeout
	}

	return DefaultStartTimeout
}
func (a *App) SetStartTimeout(startTimeout time.Duration) { a.startTimeout = startTimeout }

func (a *App) GetStopTimeout() time.Duration {
	if a.stopTimeout > 0 {
		return a.stopTimeout
	}

	return DefaultStopTimeout
}
func (a *App) SetStopTimeout(stopTimeout time.Duration) { a.stopTimeout = stopTimeout }
