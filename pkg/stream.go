package bhlnames

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"golang.org/x/sync/errgroup"
)

func (bn bhlnames) NameRefsStream(
	ctx context.Context,
	chIn <-chan input.Input,
	chOut chan<- *bhl.RefsByName,
) error {

	g, ctx := errgroup.WithContext(ctx)
	var wg sync.WaitGroup
	wg.Add(bn.cfg.JobsNum)

	for range bn.cfg.JobsNum {
		g.Go(func() error {
			defer wg.Done()
			err := bn.workerNameRefs(ctx, chIn, chOut)
			if err != nil {
				slog.Error("error in workerNameRefs", "error", err)
			}
			return err
		})
	}

	wg.Wait()
	close(chOut)

	if err := g.Wait(); err != nil {
		slog.Error("error in goroutines", "error", err)
		return err
	}

	return nil
}

func (bn bhlnames) workerNameRefs(
	ctx context.Context,
	chIn <-chan input.Input,
	chOut chan<- *bhl.RefsByName,
) error {
	for {
		select {
		case <-ctx.Done():
			for range chIn {
			}
			return ctx.Err()
		case in, ok := <-chIn:
			if !ok {
				return nil
			}
			res, err := bn.NameRefs(in)
			if err != nil {
				slog.Error("Cannot get references", "error", err)
				return err
			}
			select {
			case <-ctx.Done():
				for range chIn {
				}
				return ctx.Err()
			case chOut <- res:
			}
		}
	}
}
