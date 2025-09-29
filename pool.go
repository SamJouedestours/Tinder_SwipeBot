package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/you/swipebot/internal/api"
	"github.com/you/swipebot/internal/logic"
	"github.com/you/swipebot/internal/obs"
)

type Job struct {
	Candidate api.Candidate
}

type Result struct {
	ID      string
	Action  string
	Matched bool
	Err     error
}

type Pool struct {
	client     *api.Client
	workers    int
	jobsCh     chan Job
	resultsCh  chan Result
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

func NewPool(client *api.Client, workers int) *Pool {
	return &Pool{
		client:    client,
		workers:   workers,
		jobsCh:    make(chan Job, 256),
		resultsCh: make(chan Result, 256),
	}
}

func (p *Pool) Results() <-chan Result { return p.resultsCh }

func (p *Pool) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p.cancelFunc = cancel

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func(id int) {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-p.jobsCh:
					if !ok {
						return
					}
				action := logic.Decide(job.Candidate)
				obs.SwipeTotal.WithLabelValues(action).Inc()
				var matched bool
				err := obs.Instrument(func() error {
					resp, err := p.client.Swipe(ctx, job.Candidate.ID, action)
					if err != nil {
						return err
					}
					matched = resp.Matched
					return nil
				})
				if err != nil {
					obs.SwipeErrors.Inc()
					p.resultsCh <- Result{ID: job.Candidate.ID, Action: action, Err: err}
					continue
				}
				if matched {
					obs.MatchesTotal.Inc()
				}
				p.resultsCh <- Result{ID: job.Candidate.ID, Action: action, Matched: matched}
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(i)
	}
}

func (p *Pool) Stop() {
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	close(p.jobsCh)
	p.wg.Wait()
	close(p.resultsCh)
}

func (p *Pool) EnqueueMany(cands []api.Candidate) {
	for _, c := range cands {
		select {
		default:
			log.Printf("jobs buffer plein, skip candidate %s", c.ID)
		case p.jobsCh <- Job{Candidate: c}:
		}
	}
}
