package micro

import (
	"github.com/fabriqs/go-micro/di"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
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
		_ = di.Inject(func(r Router) {
			_ = r.Start("0.0.0.0:" + addr)
		})
	}()

	// run the cleanup after the server is terminated
	defer func() {
		_ = di.Inject(func(r Router) {
			_ = r.Shutdown()
		})

		_ = di.Inject(func(db DataSource) {
			if db != nil {
				db.Close()
			}
		})
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
	_ = di.Inject(func(db DataSource) {
		db.Close()
	})
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
