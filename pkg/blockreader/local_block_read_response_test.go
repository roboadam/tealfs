package blockreader

import (
	"context"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"testing"

	"github.com/google/uuid"
)

func TestLocalBlockReadResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inDisks := make(chan *disk.Disk)
	localReadResponses := make(chan GetFromDiskResp)
	sends := make(chan model.MgrConnsSend)
	nodeConnMap := model.NewNodeConnectionMapper()
	localNodeId := model.NewNodeId()
	remoteNodeId := model.NewNodeId()
	blockId1 := model.NewBlockId()
	blockId2 := model.NewBlockId()
	reqId1 := model.GetBlockId(uuid.NewString())
	reqId2 := model.GetBlockId(uuid.NewString())
	lbrr := LocalBlockReadResponses{
		InDisks:            inDisks,
		LocalReadResponses: localReadResponses,
		Sends:              sends,
		NodeConnMap:        nodeConnMap,
		NodeId:             localNodeId,
	}
	go lbrr.Start(ctx)

	disk1 := disk.New(disk.NewPath("p1", &disk.MockFileOps{}), localNodeId, model.DiskId(uuid.NewString()), ctx)
	disk2 := disk.New(disk.NewPath("p2", &disk.MockFileOps{}), localNodeId, model.DiskId(uuid.NewString()), ctx)
	inDisks <- &disk1
	inDisks <- &disk2

	disk1.InReads <- model.ReadRequest{
		Caller: localNodeId,
		Ptrs: []model.DiskPointer{{
			NodeId:   localNodeId,
			Disk:     disk1.Id(),
			FileName: string(blockId1),
		}},
		BlockId: blockId1,
		ReqId:   reqId1,
	}

	resp := <-localReadResponses
	if resp.Resp.Id != reqId1 {
		t.Error("invalid request id")
		return
	}

	disk2.InReads <- model.ReadRequest{
		Caller: remoteNodeId,
		Ptrs: []model.DiskPointer{{
			NodeId:   localNodeId,
			Disk:     disk2.Id(),
			FileName: string(blockId2),
		}},
		BlockId: blockId2,
		ReqId:   reqId2,
	}

	nodeConnMap.SetAll(model.ConnId(1), "someAddress:123", remoteNodeId)

	resp2 := <-sends
	payload := resp2.Payload
	if gfdr, ok := payload.(*GetFromDiskResp); !ok {
		t.Error("sent wrong type")
		return
	} else if gfdr.Resp.Id != reqId2 {
		t.Error("sent wrong message")
		return
	}
}
