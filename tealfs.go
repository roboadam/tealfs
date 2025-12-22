// Copyright (C) 2025 Adam Hess
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

	connsSvcConnectToNodeReq := make(chan model.ConnectToNodeReq, 1)
	connsSvcSendPayloadMsg := make(chan model.SendPayloadMsg, 1)

	/******* Disk Services ******/

	diskManagerSvc := disk.NewDisks(nodeId, globalPath, &disk.DiskFileOps{})
	diskManagerSvc.InAddDiskMsg = diskManagerSvcAddDiskMsg
	diskManagerSvc.InDiskAddedMsg = diskManagerSvcDiskAddedMsg
	diskManagerSvc.OutDiskAddedMsg = diskMsgSenderSvcDiskAddedMsg

	diskMsgSenderSvc := disk.MsgSenderSvc{
		InAddDiskMsg:   diskMsgSenderSvcAddDiskMsg,
		InDiskAddedMsg: diskMsgSenderSvcDiskAddedMsg,

		OutRemote: connsSvcSendPayloadMsg,

		NodeId:      nodeId,
		NodeConnMap: nodeConnMapper,
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

	/****** Startup ******/

	go diskManagerSvc.Start(ctx)
	go diskMsgSenderSvc.Start(ctx)

	/*********************/

	iamConnIds := make(chan conns.IamConnId, 1)
	incomingSyncNodes := make(chan model.SyncNodes, 1)
	sendSyncNodes := make(chan struct{}, 1)
	saveCluster := make(chan struct{}, 1)

	connsSvc.OutIamConnId = iamConnIds
	connsSvc.OutSyncNodes = incomingSyncNodes

	connsIamReceiver := conns.IamReceiver{
		InIam:            iamConnIds,
		OutSendSyncNodes: sendSyncNodes,
		OutSaveCluster:   saveCluster,
		Mapper:           nodeConnMapper,
	}
	go connsIamReceiver.Start(ctx)

	clusterSaver := conns.ClusterSaver{
		Save:           saveCluster,
		NodeConnMapper: nodeConnMapper,
		SavePath:       globalPath,
		FileOps:        &disk.DiskFileOps{},
	}
	go clusterSaver.Start(ctx)

	receiveSyncNodes := conns.ReceiveSyncNodes{
		InSyncNodes:    incomingSyncNodes,
		OutConnectTo:   connsSvcConnectToNodeReq,
		NodeConnMapper: nodeConnMapper,
		NodeId:         nodeId,
	}
	go receiveSyncNodes.Start(ctx)

	sendSyncNodesProc := conns.SendSyncNodes{
		InSendSyncNodes: sendSyncNodes,
		OutSendPayloads: connsSvcSendPayloadMsg,
		NodeConnMapper:  nodeConnMapper,
	}
	go sendSyncNodesProc.Start(ctx)

	clusterLoader := conns.ClusterLoader{
		NodeConnMapper: nodeConnMapper,
		FileOps:        &disk.DiskFileOps{},
		SavePath:       globalPath,
	}
	go clusterLoader.Load(ctx)

	reconnector := conns.Reconnector{
		OutConnectTo: connsSvcConnectToNodeReq,
		Mapper:       nodeConnMapper,
	}
	go reconnector.Start(ctx)

	newAddDiskMsgs := make(chan model.AddDiskMsg)
	u := ui.NewUi(
		connsSvcConnectToNodeReq,
		newAddDiskMsgs,
		make(chan model.UiDiskStatus),
		&ui.HttpHtmlOps{},
		nodeId,
		uiAddress,
		ctx,
	)
	u.NodeConnMap = nodeConnMapper

	addedDisksSaver := make(chan *disk.Disk)
	addedDisksReader := make(chan *disk.Disk)

	iams := make(chan model.IAm)
	connsSvc.OutIam = iams

	sendIam := make(chan model.ConnId, 1)
	connsSvc.OutSendIam = sendIam
	connsIamSender := conns.IamSender{
		InSendIam: sendIam,
		OutIam:    connsSvcSendPayloadMsg,
		NodeId:    nodeId,
		Address:   nodeAddress,
		Disks:     disks.AllDiskIds,
	}
	go connsIamSender.Start(ctx)

	webdavPutReq := make(chan model.PutBlockReq)
	webdavPutResp := make(chan model.PutBlockResp)

	localSave := make(chan blocksaver.SaveToDiskReq)
	remoteSave := make(chan blocksaver.SaveToDiskReq)
	saveResp := make(chan blocksaver.SaveToDiskResp)

	connsSvc.OutSaveToDiskReq = localSave
	connsSvc.OutSaveToDiskResp = saveResp

	bs := blocksaver.BlockSaver{
		Req:         webdavPutReq,
		RemoteDest:  remoteSave,
		LocalDest:   localSave,
		InResp:      saveResp,
		Resp:        webdavPutResp,
		Distributer: &disks.Distributer,
		NodeId:      nodeId,
	}
	go bs.Start(ctx)

	lbs := blocksaver.LocalBlockSaver{
		Req:   localSave,
		Disks: &disks.Disks,
	}
	go lbs.Start(ctx)

	rbs := blocksaver.RemoteBlockSaver{
		Req:         remoteSave,
		Sends:       connsSvcSendPayloadMsg,
		NoConnResp:  saveResp,
		NodeConnMap: nodeConnMapper,
	}
	go rbs.Start(ctx)

	lbsr := blocksaver.LocalBlockSaveResponses{
		InDisks:             addedDisksSaver,
		LocalWriteResponses: saveResp,
		Sends:               connsSvcSendPayloadMsg,
		NodeConnMap:         nodeConnMapper,
		NodeId:              nodeId,
	}
	go lbsr.Start(ctx)

	webdavGetReq := make(chan model.GetBlockReq)
	webdavGetResp := make(chan model.GetBlockResp)
	localReadDest := make(chan blockreader.GetFromDiskReq)
	remoteReadDest := make(chan blockreader.GetFromDiskReq)
	readResp := make(chan blockreader.GetFromDiskResp)

	connsSvc.OutGetFromDiskReq = localReadDest
	connsSvc.OutGetFromDiskResp = readResp

	br := blockreader.BlockReader{
		Req:         webdavGetReq,
		RemoteDest:  remoteReadDest,
		LocalDest:   localReadDest,
		InResp:      readResp,
		Resp:        webdavGetResp,
		Distributer: &disks.Distributer,
		NodeId:      nodeId,
	}
	go br.Start(ctx)

	lbr := blockreader.LocalBlockReader{
		Req:   localReadDest,
		Disks: &disks.Disks,
	}
	go lbr.Start(ctx)

	rbr := blockreader.RemoteBlockReader{
		Req:         remoteReadDest,
		Sends:       connsSvcSendPayloadMsg,
		NoConnResp:  readResp,
		NodeConnMap: nodeConnMapper,
	}
	go rbr.Start(ctx)

	lbrr := blockreader.LocalBlockReadResponses{
		InDisks:            addedDisksReader,
		LocalReadResponses: readResp,
		Sends:              connsSvcSendPayloadMsg,
		NodeConnMap:        nodeConnMapper,
		NodeId:             nodeId,
	}
	go lbrr.Start(ctx)

	inFileBroadcasts := make(chan webdav.FileBroadcast, 1)
	connsSvc.OutFileBroadcasts = inFileBroadcasts
	_ = webdav.New(
		nodeId,
		webdavGetReq,
		webdavPutReq,
		webdavGetResp,
		webdavPutResp,
		connsSvcSendPayloadMsg,
		inFileBroadcasts,
		webdavAddress,
		ctx,
		&disk.DiskFileOps{},
		globalPath,
		chansize,
		nodeConnMapper,
	)

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
