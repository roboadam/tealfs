package balancer

import (
	"context"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type Balancer struct {
	allBlocks   set.Set[model.BlockId]
	addBlock    chan model.BlockId
	removeBlock chan model.BlockId
	ctx         context.Context
}

func New(addBlock chan model.BlockId, removeBlock chan model.BlockId) *Balancer {
	b := Balancer{
		allBlocks:   set.NewSet[model.BlockId](),
		addBlock:    addBlock,
		removeBlock: removeBlock,
	}
	go b.processChannels()
	return &b
}

func (b *Balancer) processChannels() {
	for {
		select {
		case <-b.ctx.Done():
			return
		case blockId := <-b.addBlock:
			b.allBlocks.Add(blockId)
		case blockId := <-b.removeBlock:
			b.allBlocks.Remove(blockId)
		}
	}
}
