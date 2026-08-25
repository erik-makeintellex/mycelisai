package swarm

import (
	"log"
	"time"
)

func (t *Team) publishResponse(subject string, payload []byte) bool {
	if err := t.nc.Publish(subject, payload); err != nil {
		log.Printf("Team [%s] failed to publish response to [%s]: %v", t.Manifest.Name, subject, err)
		return false
	}
	return true
}

func (t *Team) flushResponses() {
	if err := t.nc.FlushTimeout(5 * time.Second); err != nil {
		log.Printf("Team [%s] failed to flush response delivery: %v", t.Manifest.Name, err)
	}
}
