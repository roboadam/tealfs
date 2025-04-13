package chanutil_test

import (
	"context"
	"tealfs/pkg/chanutil"
	"testing"
)

func TestSend(t *testing.T) {
	chanI := make(chan int)
	go chanutil.Send(context.Background(), chanI, 1, "test")
	<-chanI
}
