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
	"tealfs/pkg/blocksaver"
	"tealfs/pkg/conns"
	"tealfs/pkg/datalayer"
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
	diskIamReceiverChan := make(chan model.IAm, 1)
	diskDeleteBlocksDeleteBlockIid := make(chan disk.DeleteBlockId, 1)

	connsSvcConnectToNodeReq := make(chan model.ConnectToNodeReq, 1)
	connsIamTrigger := make(chan struct{}, 1)
	connsSendSyncNodes := make(chan struct{}, 1)
	connsClusterSaver := make(chan struct{}, 1)
	connsReceiveSyncNodes := make(chan model.SyncNodes, 1)
	localBlockSaveResponsesWriteResults := make(chan (<-chan model.WriteResult), 1)
	localBlockReadResponsesReadResults := make(chan (<-chan model.ReadResult), 1)
	blockSaverPutBlockReq := make(chan model.PutBlockReq)
	webdavPutResp := make(chan model.PutBlockResp)
	blockSaverSaveToDiskResp := make(chan blocksaver.SaveToDiskResp)
	webdavFileBroadcast := make(chan webdav.FileBroadcast, 1)
	connsPayload := make(chan model.Payload2, 1)
	webdavFetchBlockResp := make(chan model.FetchBlockResp, 1)

	/******* State Handler ******/
	stateHandler := datalayer.StateHandler{
		OutPayload:  connsPayload,
		MyNodeId:    nodeId,
		NodeConnMap: nodeConnMapper,
	}

	/******* Disk Services ******/

	diskManagerSvc := disk.NewDisks(nodeId, globalPath, &disk.DiskFileOps{})
	diskManagerSvc.InAddDiskMsg = diskManagerSvcAddDiskMsg
	diskManagerSvc.InDiskAddedMsg = diskManagerSvcDiskAddedMsg
	diskManagerSvc.OutPayload = connsPayload
	diskManagerSvc.OutAddedWriteResults = localBlockSaveResponsesWriteResults
	diskManagerSvc.OutAddedReadResults = localBlockReadResponsesReadResults

	diskDeleteBlocks := disk.DeleteBlocks{
		InDelete: diskDeleteBlocksDeleteBlockIid,
		Disks:    &diskManagerSvc.LocalDiskSvcList,
	}
	diskIamReceiver := disk.IamReceiver{
		InIam:           diskIamReceiverChan,
		OutDiskAddedMsg: diskManagerSvcDiskAddedMsg,
	}

	/****** Ui ******/

	u := ui.NewUi(
		connsSvcConnectToNodeReq,
		connsPayload,
		make(chan model.UiDiskStatus),
		&ui.HttpHtmlOps{},
		nodeId,
		uiAddress,
		ctx,
	)
	u.NodeConnMap = nodeConnMapper
	u.StateHandler = &stateHandler

	/****** BlockSaver *****/

	bs := blocksaver.BlockSaver{
		Req:          blockSaverPutBlockReq,
		InResp:       blockSaverSaveToDiskResp,
		Resp:         webdavPutResp,
		NodeId:       nodeId,
		DiskInfoList: &diskManagerSvc.DiskInfoList,
		OutPayload:   connsPayload,
	}
	lbs := blocksaver.LocalBlockSaver{
		Disks: &diskManagerSvc.LocalDiskSvcList,
	}
	lbsr := blocksaver.LocalBlockSaveResponses{
		InWriteResults:      localBlockSaveResponsesWriteResults,
		LocalWriteResponses: blockSaverSaveToDiskResp,
		NodeConnMap:         nodeConnMapper,
		NodeId:              nodeId,
		StateHandler:        &stateHandler,
		OutPayload:          connsPayload,
	}

	/******* Connection Services ******/

	connsSvc := conns.NewConns(
		connsSvcConnectToNodeReq,
		&conns.TcpConnectionProvider{},
		nodeAddress,
		nodeId,
		ctx,
	)
	connsSvc.OutDiskAddedMsg = diskManagerSvcDiskAddedMsg
	connsSvc.OutIamTrigger = connsIamTrigger
	connsSvc.OutSyncNodes = connsReceiveSyncNodes
	connsSvc.OutIam = diskIamReceiverChan
	connsSvc.LocalBlockSaver = &lbs
	connsSvc.OutSaveToDiskResp = blockSaverSaveToDiskResp
	connsSvc.OutFileBroadcasts = webdavFileBroadcast
	connsSvc.OutFetchBlockResp = webdavFetchBlockResp
	connsSvc.StateHandler = &stateHandler
	connsSvc.InPayload = connsPayload
	connsSvc.DiskManager = diskManagerSvc
	connsSvc.OutAddDiskMsg = diskManagerSvcAddDiskMsg
	connsSvc.NodeConnMapper = nodeConnMapper
	connsIamReceiver := conns.IamReceiver{
		InIamTrigger:     connsIamTrigger,
		OutSendSyncNodes: connsSendSyncNodes,
		OutSaveCluster:   connsClusterSaver,
		Mapper:           nodeConnMapper,
	}
	connsSendSyncNodesProc := conns.SendSyncNodes{
		InSendSyncNodes: connsSendSyncNodes,
		OutPayload:      connsPayload,
		NodeConnMapper:  nodeConnMapper,
		NodeId:          nodeId,
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
	iamGenerator := conns.IamSender{
		NodeId:  nodeId,
		Address: nodeAddress,
		Disks:   &diskManagerSvc.DiskInfoList,
	}
	connsSvc.IamGenerator = &iamGenerator

	/****** Webdav *******/

	_ = webdav.New(
		nodeId,
		blockSaverPutBlockReq,
		webdavFetchBlockResp,
		webdavPutResp,
		connsPayload,
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
	go diskDeleteBlocks.Start(ctx)
	go diskIamReceiver.Start(ctx)
	go connsIamReceiver.Start(ctx)
	go connsSendSyncNodesProc.Start(ctx)
	go connsClusterSaverSvc.Start(ctx)
	go receiveSyncNodes.Start(ctx)
	go clusterLoader.Load(ctx)
	go reconnector.Start(ctx)
	go bs.Start(ctx)
	go lbsr.Start(ctx)

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
