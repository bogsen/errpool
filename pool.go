package errpool // import "go.bog.dev/errpool"

import (
	"context"
	"golang.org/x/sync/errgroup"
)

type Pool struct {
	g       *errgroup.Group
	ctx     context.Context
	workers int
	tasks   chan func(context.Context) error
	closed  bool
}

func Unbounded(ctx context.Context) *Pool {
	return WithNumWorkers(ctx, -1)
}

func WithNumWorkers(ctx context.Context, workers int) *Pool {
	g, ctx := errgroup.WithContext(ctx)
	tasks := make(chan func(context.Context) error)
	pool := &Pool{
		g:       g,
		ctx:     ctx,
		workers: workers,
		tasks:   tasks,
	}
	for i := 0; i < workers; i++ {
		g.Go(pool.worker)
	}
	return pool
}

func (p *Pool) worker() error {
	var err error
	for f := range p.tasks {
		if err == nil {
			err = f(p.ctx)
			if p.workers <= 0 {
				break
			}
		}
	}
	return err
}

func (p *Pool) Context() context.Context {
	return p.ctx
}

func (p *Pool) Go(f func(ctx context.Context) error) {
	if p.workers <= 0 {
		p.g.Go(p.worker)
	}
	p.tasks <- f
}

func (p *Pool) Wait() error {
	if !p.closed {
		p.closed = true
		close(p.tasks)
	}
	return p.g.Wait()
}
