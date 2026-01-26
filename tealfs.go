// Copyright (C) 2026 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"tealfs/pkg/blockreader"
	"tealfs/pkg/blocksaver"
	"tealfs/pkg/conns"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/ui"
	"tealfs/pkg/webdav"

	log "github.com/sirupsen/logrus"
)

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error getting user config directory:", err)
		return
	}
	configDir = filepath.Join(configDir, "tealfs")
	if err = os.Mkdir(configDir, 0700); err != nil && !errors.Is(err, fs.ErrExist) {
		fmt.Printf("unable to create config directory: {%s}. error: %s\n", configDir, err)
		os.Exit(1)
	}

	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, os.Args[0], "<webdav address> <ui address> <node address>")
		os.Exit(1)
	}

	_ = startTealFs(configDir, os.Args[1], os.Args[2], os.Args[3], context.Background())
}

func startTealFs(globalPath string, webdavAddress string, uiAddress string, nodeAddress string, ctx context.Context) error {
	log.SetLevel(log.DebugLevel)
	chansize := 0
	nodeConnMapper := model.NewNodeConnectionMapper()

	nodeId, err := readNodeId(globalPath, &disk.DiskFileOps{})
	if err != nil {
		log.Fatal("Unable to read Node Id")
	}

	/***** Channels, name starts with receiving service **********/

	diskManagerSvcDiskAddedMsg := make(chan model.DiskAddedMsg, 1)
	diskManagerSvcAddDiskMsg := make(chan model.AddDiskMsg, 1)
	diskMsgSenderSvcDiskAddedMsg := make(chan model.DiskAddedMsg, 1)
	diskMsgSenderSvcAddDiskMsg := make(chan model.AddDiskMsg, 1)
	diskIamReceiverChan := make(chan model.IAm, 1)
	diskDeleteBlocksDeleteBlockIid := make(chan disk.DeleteBlockId, 1)

	connsSvcConnectToNodeReq := make(chan model.ConnectToNodeReq, 1)
	connsSvcSendPayloadMsg := make(chan model.SendPayloadMsg, 1)
	connsIamReceiverIamConnId := make(chan conns.IamConnId, 1)
	connsSendSyncNodes := make(chan struct{}, 1)
	connsClusterSaver := make(chan struct{}, 1)
	connsReceiveSyncNodes := make(chan model.SyncNodes, 1)
	connsIamSenderConnId := make(chan model.ConnId, 1)
	localBlockSaveResponsesWriteResults := make(chan (<-chan model.WriteResult), 1)
	localBlockSaverSaveToDiskReq := make(chan blocksaver.SaveToDiskReq)
	localBlockReadResponsesReadResults := make(chan (<-chan model.ReadResult), 1)
	blockSaverPutBlockReq := make(chan model.PutBlockReq)
	webdavPutResp := make(chan model.PutBlockResp)
	remoteBlockSaverSaveToDiskReq := make(chan blocksaver.SaveToDiskReq)
	blockSaverSaveToDiskResp := make(chan blocksaver.SaveToDiskResp)
	blockReaderGetBlockReq := make(chan model.GetBlockReq)
	webdavGetBlockResp := make(chan model.GetBlockResp)
	localBlockReaderGetFromDiskReq := make(chan blockreader.GetFromDiskReq)
	remoteBlockReaderGetFromDiskReq := make(chan blockreader.GetFromDiskReq)
	blockReaderGetFromDiskResp := make(chan blockreader.GetFromDiskResp)
	webdavFileBroadcast := make(chan webdav.FileBroadcast, 1)

	/******* Disk Services ******/

	diskManagerSvc := disk.NewDisks(nodeId, globalPath, &disk.DiskFileOps{})
	diskManagerSvc.InAddDiskMsg = diskManagerSvcAddDiskMsg
	diskManagerSvc.InDiskAddedMsg = diskManagerSvcDiskAddedMsg
	diskManagerSvc.OutDiskAddedMsg = diskMsgSenderSvcDiskAddedMsg
	diskManagerSvc.OutAddedWriteResults = localBlockSaveResponsesWriteResults
	diskManagerSvc.OutAddedReadResults = localBlockReadResponsesReadResults

	diskMsgSenderSvc := disk.MsgSenderSvc{
		InAddDiskMsg:   diskMsgSenderSvcAddDiskMsg,
		InDiskAddedMsg: diskMsgSenderSvcDiskAddedMsg,
		OutRemote:      connsSvcSendPayloadMsg,
		NodeId:         nodeId,
		NodeConnMap:    nodeConnMapper,
	}
	diskDeleteBlocks := disk.DeleteBlocks{
		InDelete: diskDeleteBlocksDeleteBlockIid,
		Disks:    &diskManagerSvc.LocalDiskSvcList,
	}
	diskIamReceiver := disk.IamReceiver{
		InIam:           diskIamReceiverChan,
		OutDiskAddedMsg: diskManagerSvcDiskAddedMsg,
	}

	/******* Connection Services ******/

	connsSvc := conns.NewConns(
		connsSvcConnectToNodeReq,
		connsSvcSendPayloadMsg,
		&conns.TcpConnectionProvider{},
		nodeAddress,
		nodeId,
		ctx,
	)
	connsSvc.OutDiskAddedMsg = diskManagerSvcDiskAddedMsg
	connsSvc.OutIamConnId = connsIamReceiverIamConnId
	connsSvc.OutSyncNodes = connsReceiveSyncNodes
	connsSvc.OutIam = diskIamReceiverChan
	connsSvc.OutSendIam = connsIamSenderConnId
	connsSvc.OutSaveToDiskReq = localBlockSaverSaveToDiskReq
	connsSvc.OutSaveToDiskResp = blockSaverSaveToDiskResp
	connsSvc.OutGetFromDiskReq = localBlockReaderGetFromDiskReq
	connsSvc.OutGetFromDiskResp = blockReaderGetFromDiskResp
	connsSvc.OutFileBroadcasts = webdavFileBroadcast
	connsIamReceiver := conns.IamReceiver{
		InIam:            connsIamReceiverIamConnId,
		OutSendSyncNodes: connsSendSyncNodes,
		OutSaveCluster:   connsClusterSaver,
		Mapper:           nodeConnMapper,
	}
	connsSendSyncNodesProc := conns.SendSyncNodes{
		InSendSyncNodes: connsSendSyncNodes,
		OutSendPayloads: connsSvcSendPayloadMsg,
		NodeConnMapper:  nodeConnMapper,
	}
	connsClusterSaverSvc := conns.ClusterSaver{
		Save:           connsClusterSaver,
		NodeConnMapper: nodeConnMapper,
		SavePath:       globalPath,
		FileOps:        &disk.DiskFileOps{},
	}
	receiveSyncNodes := conns.ReceiveSyncNodes{
		InSyncNodes:  connsReceiveSyncNodes,
		OutConnectTo: connsSvcConnectToNodeReq,

		NodeConnMapper: nodeConnMapper,
		NodeId:         nodeId,
	}
	clusterLoader := conns.ClusterLoader{
		NodeConnMapper: nodeConnMapper,
		FileOps:        &disk.DiskFileOps{},
		SavePath:       globalPath,
	}
	reconnector := conns.Reconnector{
		OutConnectTo: connsSvcConnectToNodeReq,
		Mapper:       nodeConnMapper,
	}
	connsIamSender := conns.IamSender{
		InSendIam: connsIamSenderConnId,
		OutIam:    connsSvcSendPayloadMsg,
		NodeId:    nodeId,
		Address:   nodeAddress,
		Disks:     &diskManagerSvc.DiskInfoList,
	}

	/****** Ui ******/

	u := ui.NewUi(
		connsSvcConnectToNodeReq,
		diskManagerSvcAddDiskMsg,
		diskMsgSenderSvcAddDiskMsg,
		make(chan model.UiDiskStatus),
		&ui.HttpHtmlOps{},
		nodeId,
		uiAddress,
		ctx,
	)
	u.NodeConnMap = nodeConnMapper

	/****** BlockSaver *****/

	bs := blocksaver.BlockSaver{
		Req:         blockSaverPutBlockReq,
		RemoteDest:  remoteBlockSaverSaveToDiskReq,
		LocalDest:   localBlockSaverSaveToDiskReq,
		InResp:      blockSaverSaveToDiskResp,
		Resp:        webdavPutResp,
		Distributer: &diskManagerSvc.Distributer,
		NodeId:      nodeId,
	}
	lbs := blocksaver.LocalBlockSaver{
		Req:   localBlockSaverSaveToDiskReq,
		Disks: &diskManagerSvc.LocalDiskSvcList,
	}
	rbs := blocksaver.RemoteBlockSaver{
		Req:         remoteBlockSaverSaveToDiskReq,
		Sends:       connsSvcSendPayloadMsg,
		NoConnResp:  blockSaverSaveToDiskResp,
		NodeConnMap: nodeConnMapper,
	}
	lbsr := blocksaver.LocalBlockSaveResponses{
		InWriteResults:      localBlockSaveResponsesWriteResults,
		LocalWriteResponses: blockSaverSaveToDiskResp,
		Sends:               connsSvcSendPayloadMsg,
		NodeConnMap:         nodeConnMapper,
		NodeId:              nodeId,
	}

	/****** BlockReader *****/

	br := blockreader.BlockReader{
		Req:         blockReaderGetBlockReq,
		RemoteDest:  remoteBlockReaderGetFromDiskReq,
		LocalDest:   localBlockReaderGetFromDiskReq,
		InResp:      blockReaderGetFromDiskResp,
		Resp:        webdavGetBlockResp,
		Distributer: &diskManagerSvc.Distributer,
		NodeId:      nodeId,
	}
	lbr := blockreader.LocalBlockReader{
		Req:   localBlockReaderGetFromDiskReq,
		Disks: &diskManagerSvc.LocalDiskSvcList,
	}
	rbr := blockreader.RemoteBlockReader{
		Req:         remoteBlockReaderGetFromDiskReq,
		Sends:       connsSvcSendPayloadMsg,
		NoConnResp:  blockReaderGetFromDiskResp,
		NodeConnMap: nodeConnMapper,
	}
	lbrr := blockreader.LocalBlockReadResponses{
		InReadResults:      localBlockReadResponsesReadResults,
		LocalReadResponses: blockReaderGetFromDiskResp,
		Sends:              connsSvcSendPayloadMsg,
		NodeConnMap:        nodeConnMapper,
		NodeId:             nodeId,
	}

	/****** Webdav *******/

	_ = webdav.New(
		nodeId,
		blockReaderGetBlockReq,
		blockSaverPutBlockReq,
		webdavGetBlockResp,
		webdavPutResp,
		connsSvcSendPayloadMsg,
		webdavFileBroadcast,
		webdavAddress,
		ctx,
		&disk.DiskFileOps{},
		globalPath,
		chansize,
		nodeConnMapper,
	)

	/****** Startup ******/

	go diskManagerSvc.Start(ctx)
	go diskMsgSenderSvc.Start(ctx)
	go diskDeleteBlocks.Start(ctx)
	go diskIamReceiver.Start(ctx)
	go connsIamReceiver.Start(ctx)
	go connsSendSyncNodesProc.Start(ctx)
	go connsClusterSaverSvc.Start(ctx)
	go receiveSyncNodes.Start(ctx)
	go clusterLoader.Load(ctx)
	go reconnector.Start(ctx)
	go connsIamSender.Start(ctx)
	go bs.Start(ctx)
	go lbs.Start(ctx)
	go rbs.Start(ctx)
	go lbsr.Start(ctx)
	go br.Start(ctx)
	go lbr.Start(ctx)
	go rbr.Start(ctx)
	go lbrr.Start(ctx)

	<-ctx.Done()
	return nil
}

func readNodeId(savePath string, fileOps disk.FileOps) (model.NodeId, error) {
	data, err := fileOps.ReadFile(filepath.Join(savePath, "node_id"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			nodeId := model.NewNodeId()
			err = fileOps.WriteFile(filepath.Join(savePath, "node_id"), []byte(nodeId))
			if err != nil {
				return "", err
			}
			return nodeId, nil
		}
		return "", err
	}
	return model.NodeId(data), nil
}
