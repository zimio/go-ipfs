// Package provider implements structures and methods to provide blocks,
// keep track of which blocks are provided, and to allow those blocks to
// be reprovided.
package simple

import (
	"context"

	cid "github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
	q "github.com/ipfs/go-ipfs/provider/queue"
	routing "github.com/libp2p/go-libp2p-routing"
)

var plog = logging.Logger("provider/deprecated")

const provideOutgoingWorkerLimit = 8

type Provider struct {
	ctx context.Context
	// the CIDs for which provide announcements should be made
	queue *q.Queue
	// used to announce providing to the network
	contentRouting routing.ContentRouting
}

// NewProvider creates a provider that announces blocks to the network using a content router
func NewProvider(ctx context.Context, queue *q.Queue, contentRouting routing.ContentRouting) *Provider {
	return &Provider{
		ctx:            ctx,
		queue:          queue,
		contentRouting: contentRouting,
	}
}

// Close stops the provider
func (p *Provider) Close() error {
	p.queue.Close()
	return nil
}

// Start workers to handle provide requests.
func (p *Provider) Run() {
	p.handleAnnouncements()
}

// Provide the given cid using specified strategy.
func (p *Provider) Provide(root cid.Cid) error {
	p.queue.Enqueue(root)
	return nil
}

// Handle all outgoing cids by providing (announcing) them
func (p *Provider) handleAnnouncements() {
	for workers := 0; workers < provideOutgoingWorkerLimit; workers++ {
		go func() {
			for p.ctx.Err() == nil {
				select {
				case <-p.ctx.Done():
					return
				case c := <-p.queue.Dequeue():
					plog.Info("announce - start - ", c)
					if err := p.contentRouting.Provide(p.ctx, c, true); err != nil {
						plog.Warningf("Unable to provide entry: %s, %s", c, err)
					}
					plog.Info("announce - end - ", c)
				}
			}
		}()
	}
}