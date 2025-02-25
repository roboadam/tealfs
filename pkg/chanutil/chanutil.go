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
}

func checkDone(ctx context.Context, message string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
			log.Error("Send timed out:", message)
		}
	}
}

func Receive[T any](channel chan T, message string) T {
	uuid := uuid.New()
	log.Trace("R:B:", uuid, ":", message)
	value := <-channel
	log.Trace("R:A:", uuid, ":", message)
	return value
}
