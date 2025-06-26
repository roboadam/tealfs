package balancer

import (
	"context"
	"tealfs/pkg/conns"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type Balancer struct {
	allBlocks      set.Set[model.BlockId]
	addBlock       chan model.BlockId
	removeBlock    chan model.BlockId
	nodeConnMapper *model.NodeConnectionMapper
	conns          *conns.Conns
	ctx            context.Context
}

func New(ctx context.Context, nodeConnMapper *model.NodeConnectionMapper) *Balancer {
	b := Balancer{
		allBlocks:      set.NewSet[model.BlockId](),
		addBlock:       make(chan model.BlockId),
		removeBlock:    make(chan model.BlockId),
		nodeConnMapper: nodeConnMapper,
		ctx:            ctx,
	}
	go b.processChannels()
	return &b
}

func (b *Balancer) AddBlock(blockId model.BlockId) {
	b.addBlock <- blockId
	connections := b.nodeConnMapper.Connections()
	for _, connId := range connections.GetValues() {
		b.conns.Send(connId, model.NewAddBlockReq(blockId), nil)
	}
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
