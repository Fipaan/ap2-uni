package notify

import (
	"github.com/Fipaan/ap2-uni/config"
)

func NewFromEnv() Provider {
    if config.ProviderMode() == "REAL" {
        panic("REAL provider not implemented")
    }
    return NewSimulatedProvider()
}
