package core

import (
	"github.com/fabriqs/go-micro/db"
	"github.com/fabriqs/go-micro/di"
	"github.com/fabriqs/go-micro/router"
	"os"
	"os/signal"
	"syscall"
)

func (a *App) Run(addr string) {
	// setup exit code for graceful shutdown
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// start the server
	go func() {
		_ = di.Inject(func(r router.R) {
			_ = r.Start("0.0.0.0:" + addr)
		})
	}()

	// run the cleanup after the server is terminated
	defer func() {
		_ = di.Inject(func(r router.R) {
			_ = r.Shutdown()
		})

		_ = di.Inject(func(db db.DB) {
			if db != nil {
				db.Close()
			}
		})
	}()

	/*if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}*/

	// ensure the server is shutdown gracefully & app runs
	gracefully()
}

func (a *App) Cleanup() {
	_ = di.Inject(func(db db.DB) {
		db.Close()
	})
}

func NewApp(name string, features []Feature) *App {

	return &App{
		Name:     name,
		Features: features,
		//Router:   r,
	}
}

func gracefully() {
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
