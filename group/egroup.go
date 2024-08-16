package main

import (
	"context"
	"fmt"
	"github.com/songzhibin97/gkit/egroup"
	"github.com/songzhibin97/gkit/goroutine"
	"time"
)

func main() {

	ctxTimeout, cancel := context.WithCancel(context.Background())
	defer cancel()
	group := goroutine.NewGoroutine(ctxTimeout, goroutine.SetMax(3))
	contextGroup := egroup.WithContextGroup(ctxTimeout, group)
	s := make([]int, 100)
	for i := 0; i < 100; i++ {
		s[i] = i
	}
	lens := len(s)
	fmt.Println("lens=", lens)
	limit := 10
	for i := 0; i < lens; i += limit {
		end := i + limit
		if end > lens {
			end = lens
		}
		st := s[i:end]

		contextGroup.Go(func() error {
			fmt.Println("s=", st)
			time.Sleep(1 * time.Second)
			return nil
		})
	}
	contextGroup.Wait()

}
