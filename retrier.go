package golang_use_example

import (
	"context"
	"github.com/songzhibin97/gkit/egroup"
	"github.com/songzhibin97/gkit/goroutine"
	"testing"
)

func main(t *testing.T) {

	ctxTimeout, cancel := context.WithCancel(context.Background())
	defer cancel()
	group := goroutine.NewGoroutine(ctxTimeout, goroutine.SetMax(100))
	egroup.WithContextGroup(ctxTimeout, group)

}
