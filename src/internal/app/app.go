package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/CE-Thesis-2023/backend/src/internal/configs"
	custcron "github.com/CE-Thesis-2023/backend/src/internal/cron"
	custhttp "github.com/CE-Thesis-2023/backend/src/internal/http"
	"github.com/CE-Thesis-2023/backend/src/internal/logger"

	"go.uber.org/zap"
)

func Run(shutdownTimeout time.Duration, registration RegistrationFunc) {
	ctx := context.Background()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT)

	configs.Init(ctx)

	globalConfigs := configs.Get()

	loggerConfigs := globalConfigs.Logger
	logger.Init(ctx, logger.WithGlobalConfigs(&loggerConfigs))

	options := registration(globalConfigs, logger.Logger())

	opts := Options{}
	for _, optioner := range options {
		optioner(&opts)
	}

	go func() {
		logger := zap.L().Sugar()
		logger.Infof("Run: configs = %s", globalConfigs.String())

		if opts.factoryHook != nil {
			if err := opts.factoryHook(); err != nil {
				logger.Errorf("Run: factoryHook err = %s", err)
				quit <- syscall.SIGTERM
				return
			}
		}

		for _, s := range opts.httpServers {
			s := s
			go func() {
				logger.Infof("Run: start HTTP server name = %s", s.Name())
				if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Infof("Run: start HTTP server err = %s", err)
					quit <- syscall.SIGTERM
				}
			}()
		}

		for _, s := range opts.schedulers {
			s := s
			go func() {
				logger.Infof("Run: start scheduler name = %s", s.Name())
				if err := s.Start(); err != nil {
					logger.Infof("Run: start scheduler err = %s", err)
					quit <- syscall.SIGTERM
				}
			}()
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if opts.shutdownHook != nil {
		opts.shutdownHook(ctx)
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		for _, s := range opts.httpServers {
			s := s
			logger.Infof("Run: stop HTTP server name = %s", s.Name())
			if err := s.Stop(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, s := range opts.schedulers {
			s := s
			logger.Infof("Run: stop scheduler name = %s", s.Name())
			if err := s.Stop(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}()

	wg.Wait()

	zap.L().Sync()
	log.Print("Run: shutdown complete")
}

type RegistrationFunc func(configs *configs.Configs, logger *zap.Logger) []Optioner
type FactoryHook func() error
type ShutdownHook func(ctx context.Context)

type Options struct {
	httpServers []*custhttp.HttpServer
	schedulers  []*custcron.Scheduler

	factoryHook  FactoryHook
	shutdownHook ShutdownHook
}

type Optioner func(opts *Options)

func WithHttpServer(server *custhttp.HttpServer) Optioner {
	return func(opts *Options) {
		if server != nil {
			opts.httpServers = append(opts.httpServers, server)
		}
	}
}

func WithFactoryHook(cb FactoryHook) Optioner {
	return func(opts *Options) {
		opts.factoryHook = cb
	}
}

func WithShutdownHook(cb ShutdownHook) Optioner {
	return func(opts *Options) {
		opts.shutdownHook = cb
	}
}

func WithScheduling(scheduler *custcron.Scheduler) Optioner {
	return func(opts *Options) {
		opts.schedulers = append(opts.schedulers, scheduler)
	}
}
