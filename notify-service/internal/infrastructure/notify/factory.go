package notify

import (
	"github.com/Fipaan/ap2-uni/config"
)

func NewFromEnv() Provider {
    if config.ProviderMode() == "REAL" {
        return NewSMTPProvider()
    }
    return NewSimulatedProvider()
}
