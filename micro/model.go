package micro

import (
	"github.com/fabriqs/go-micro/schema"
)

type Ctx interface {
	Config() interface{}
	DB() DataSource

	Auth() *schema.Authentication
	IsAuthenticated() bool
}
