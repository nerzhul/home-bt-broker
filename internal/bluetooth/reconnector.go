package bluetooth

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"
)

// StartReconnectLoop periodically attempts to connect to paired+trusted devices not connected
func StartReconnectLoop(ctx context.Context, mgr BluetoothManagerInterface, interval time.Duration) {
	go func() {
		log.Printf("Reconnect: loop started (interval=%s)", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Reconnect: loop stopped")
				return
			case <-ticker.C:
				adapters, err := mgr.GetAdapters()
				if err != nil {
					log.Printf("Reconnect: get adapters error: %v", err)
					continue
				}
				for _, a := range adapters {
					devices, err := mgr.GetDevices(a.Path)
					if err != nil {
						log.Printf("Reconnect: get devices error for %s: %v", a.Address, err)
						continue
					}
					for _, d := range devices {
						if d.Paired && d.Trusted && !d.Connected {
							// Attempt connection with shorter timeout inside ConnectDevice
							if err := mgr.ConnectDevice(a.Path, d.Address); err != nil {
								log.Printf("Reconnect: connect %s failed: %v", d.Address, err)
							} else {
								log.Printf("Reconnect: connect attempt for %s initiated", d.Address)
							}
						}
					}
				}
			}
		}
	}()
}

// GetReconnectInterval returns interval from env RECONNECT_INTERVAL_SECONDS or default seconds
func GetReconnectInterval(defaultSeconds int) time.Duration {
	if v := os.Getenv("RECONNECT_INTERVAL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(defaultSeconds) * time.Second
}
