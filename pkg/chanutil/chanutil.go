package chanutil

import (
	"context"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func Send[T any](ctx context.Context, channel chan T, value T, message string) {
	checkDoneCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go checkDone(checkDoneCtx, message)
	select {
	case <-ctx.Done():
		return
	case channel <- value:
	}
}

func checkDone(ctx context.Context, message string) {
	timedOut := false
	uuid := uuid.New().String()[:8]
	for {
		select {
		case <-ctx.Done():
			if timedOut {
				log.Info("SND:SUC:", uuid, ":", message)
			}
			return
		case <-time.After(time.Second * 5):
			log.Warn("SND:TIM:", uuid, ":", message)
			timedOut = true
		}
	}
}
