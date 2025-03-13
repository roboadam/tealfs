package chanutil

import (
	"context"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func Send[T any](channel chan T, value T, message string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go checkDone(ctx, message)
	channel <- value
	cancel()
}

func checkDone(ctx context.Context, message string) {
	timedOut := false
	uuid := uuid.New().String()[:8]
	for {
		select {
		case <-ctx.Done():
			if timedOut {
				log.Error("SND:SUC:", uuid, ":", message)
			}
			return
		case <-time.After(time.Second * 5):
			log.Error("SND:TIM:", uuid, ":", message)
			timedOut = true
		}
	}
}
