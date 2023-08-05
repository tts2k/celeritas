package middleware

import (
	"github.com/tts2k/celeritas"

	"myapp/data"
)

type Middleware struct {
	App *celeritas.Celeritas
	Models data.Models
}
