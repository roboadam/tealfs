package chanutil_test

import (
	"tealfs/pkg/chanutil"
	"testing"
)

func TestSend(t *testing.T) {
	chanI := make(chan int)
	go chanutil.Send(chanI, 1, "test")
	<-chanI
}
