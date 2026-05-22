package internal

import (
	"time"
)

func schedulePurge() {
	go func() {
		for {
			PurgeOldRecords()
			time.Sleep(24 * time.Hour)
		}
	}()
}
