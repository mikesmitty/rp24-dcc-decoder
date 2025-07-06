//go:build rp

package hal

import (
	"machine"
)

func watchdogInit() {
	config := machine.WatchdogConfig{
		TimeoutMillis: 1000,
	}
	machine.Watchdog.Configure(config)

	machine.Watchdog.Start()
}

func WatchdogReset() {
	machine.Watchdog.Update()
}
