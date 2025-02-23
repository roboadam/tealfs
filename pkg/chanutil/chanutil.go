package chanutil

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func Send[T any](channel chan T, value T, message string) {
	uuid := uuid.New()
	log.Trace("S:B:", uuid, ":", message)
	channel <- value
	log.Trace("S:A:", uuid, ":", message)
}

func Receive[T any](channel chan T, message string) T {
	uuid := uuid.New()
	log.Trace("R:B:", uuid, ":", message)
	value := <-channel
	log.Trace("R:A:", uuid, ":", message)
	return value
}
