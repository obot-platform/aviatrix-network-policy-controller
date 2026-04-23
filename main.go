package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/stdr"
	"github.com/obot-platform/aviatrix-network-policy-controller/pkg/app"
	ctrlruntimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	log.SetFlags(log.LstdFlags)
	ctrlruntimelog.SetLogger(stdr.New(log.Default()))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	for {
		if err := app.Run(ctx); err != nil {
			if ctx.Err() != nil {
				break
			}

			fmt.Fprintf(os.Stderr, "aviatrix-network-policy-controller: %v\n", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		return
	}

	os.Exit(1)
}
