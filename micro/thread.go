package micro

import (
	"github.com/fabriqs/go-micro/schema"
	log "github.com/sirupsen/logrus"
	"github.com/timandy/routine"
)

// var ThreadLocal = routine.NewThreadLocal()
var currentContext = routine.NewInheritableThreadLocal()

func SetContext(ctx Ctx) {
	currentContext.Set(ctx)
}

func ClearContext() {
	currentContext.Remove()
}

func CurrentContext() Ctx {
	ctx := currentContext.Get()
	return ctx.(Ctx)
}

func HasContext() bool {
	return currentContext.Get() != nil
}

func CurrentUser() *schema.Authentication {
	ctx := CurrentContext()
	if ctx == nil {
		log.Fatal("NO_CONTEXT_PROVIDED")
	}
	return CurrentContext().Auth()
}

func currentDB() DataSource {
	ctx := CurrentContext()
	if ctx == nil {
		log.Fatal("NO_CONTEXT_PROVIDED")
	}
	return ctx.DB()
}
