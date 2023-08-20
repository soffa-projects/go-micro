package micro

import (
	log "github.com/sirupsen/logrus"
	"github.com/timandy/routine"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func (a *App) Run(addr string) {
	// setup exit code for graceful shutdown
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// start the server
	go func() {
		_ = a.Router.Start("0.0.0.0:" + addr)
	}()

	// run the cleanup after the server is terminated
	defer func() {
		_ = a.Router.Shutdown()
		if a.Env.DataSource != nil {
			a.Env.DataSource.Close()
		}
	}()

	if a.Scheduler != nil {
		go func() {
			time.Sleep(5 * time.Second)
			a.Scheduler.StartAsync()
		}()
	}
	/*if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}*/

	// ensure the server is shutdown gracefully & app runs
	gracefully()
}

func (a *App) Cleanup() {
	if a.Env.DataSource != nil {
		a.Env.DataSource.Close()
	}
}

func (a *App) Init() *App {
	for _, feat := range a.Features {
		var err error
		if feat.Init != nil {
			err = feat.Init(a)
		}
		if err != nil {
			log.Fatalf("failed to init feature %s.\n%v", feat.Name, err)
			return nil
		}
	}
	return a
}

func gracefully() {
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func Invoke(ctx Ctx, cb func() (interface{}, error)) (out interface{}, err error) {
	SetContext(ctx)
	defer func() {
		ClearContext()
	}()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	routine.Go(func() {
		defer func() {
			if err0 := recover(); err0 != nil {
				err = err0.(error)
			}
			wg.Done()
		}()
		out, err = cb()
	})
	wg.Wait()
	return out, err
}
