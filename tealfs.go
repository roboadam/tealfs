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
	connReqs := make(chan model.ConnectToNodeReq, 1)
	nodeConnMapper := model.NewNodeConnectionMapper()

	nodeId, err := readNodeId(globalPath, &disk.DiskFileOps{})
	if err != nil {
		log.Fatal("Unable to read Node Id")
	}

	iamConnIds := make(chan conns.IamConnId, 1)
	incomingSyncNodes := make(chan model.SyncNodes, 1)
	sendSyncNodes := make(chan struct{}, 1)
	saveCluster := make(chan struct{}, 1)
	netSends := make(chan model.MgrConnsSend, 1)

	connsMain := conns.NewConns(
		connReqs,
		netSends,
		&conns.TcpConnectionProvider{},
		nodeAddress,
		nodeId,
		ctx,
	)

	connsMain.OutIamConnId = iamConnIds
	connsMain.OutSyncNodes = incomingSyncNodes

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
		OutConnectTo:   connReqs,
		NodeConnMapper: nodeConnMapper,
		NodeId:         nodeId,
	}
	go receiveSyncNodes.Start(ctx)

	sendSyncNodesProc := conns.SendSyncNodes{
		InSendSyncNodes: sendSyncNodes,
		OutSendPayloads: netSends,
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
		OutConnectTo: connReqs,
		Mapper:       nodeConnMapper,
	}
	go reconnector.Start(ctx)

	newAddDiskMsgs := make(chan model.AddDiskMsg)
	u := ui.NewUi(
		connReqs,
		newAddDiskMsgs,
		make(chan model.UiDiskStatus),
		&ui.HttpHtmlOps{},
		nodeId,
		uiAddress,
		ctx,
	)
	u.NodeConnMap = nodeConnMapper

	localAddDiskMsgs := make(chan model.AddDiskMsg, 1)
	diskMsgSender := disk.MsgSenderSvc{
		InAddDiskMsg:  newAddDiskMsgs,
		OutAddDiskMsg: localAddDiskMsgs,
		OutRemote:     netSends,
		NodeId:        nodeId,
		NodeConnMap:   nodeConnMapper,
	}
	go diskMsgSender.Start(ctx)

	addedDisksSaver := make(chan *disk.Disk)
	addedDisksReader := make(chan *disk.Disk)

	localAddDiskReqs := make(chan model.AddDiskReq)
	remoteAddDiskReqs := make(chan model.AddDiskReq)

	connsMain.OutAddDiskReq = newAddDiskReqs
	disks := disk.NewDisks(nodeId, globalPath, &disk.DiskFileOps{})
	disks.InAddDiskReq = newAddDiskReqs
	disks.AllDiskIds.OutDiskAdded = newAddDiskReqs
	disks.OutLocalAddDiskReq = localAddDiskReqs
	disks.OutRemoteAddDiskReq = remoteAddDiskReqs
	go disks.Start(ctx)

	iamDiskUpdates := make(chan []model.AddDiskReq, 1)

	localDiskAdder := disk.LocalDiskAdder{
		InAddDiskReq: localAddDiskReqs,
		OutAddLocalDisk: []chan<- *disk.Disk{
			addedDisksSaver,
			addedDisksReader,
		},
		OutIamDiskUpdate: iamDiskUpdates,
		FileOps:          &disk.DiskFileOps{},
		Disks:            &disks.Disks,
		Distributer:      &disks.Distributer,
		AllDiskIds:       disks.AllDiskIds,
	}
	go localDiskAdder.Start(ctx)

	iamSender := disk.IamSender{
		InIamDiskUpdate: iamDiskUpdates,
		OutSends:        netSends,
		Mapper:          nodeConnMapper,
		NodeId:          nodeId,
		Address:         webdavAddress,
	}
	go iamSender.Start(ctx)

	iams := make(chan model.IAm)
	connsMain.OutIam = iams
	iamReceiver := disk.IamReceiver{
		InIam:       iams,
		Distributer: &disks.Distributer,
		AllDiskIds:  disks.AllDiskIds,
	}
	go iamReceiver.Start(ctx)

	sendIam := make(chan model.ConnId, 1)
	connsMain.OutSendIam = sendIam
	connsIamSender := conns.IamSender{
		InSendIam: sendIam,
		OutIam:    netSends,
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

	connsMain.OutSaveToDiskReq = localSave
	connsMain.OutSaveToDiskResp = saveResp

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
		Sends:       netSends,
		NoConnResp:  saveResp,
		NodeConnMap: nodeConnMapper,
	}
	go rbs.Start(ctx)

	lbsr := blocksaver.LocalBlockSaveResponses{
		InDisks:             addedDisksSaver,
		LocalWriteResponses: saveResp,
		Sends:               netSends,
		NodeConnMap:         nodeConnMapper,
		NodeId:              nodeId,
	}
	go lbsr.Start(ctx)

	webdavGetReq := make(chan model.GetBlockReq)
	webdavGetResp := make(chan model.GetBlockResp)
	localReadDest := make(chan blockreader.GetFromDiskReq)
	remoteReadDest := make(chan blockreader.GetFromDiskReq)
	readResp := make(chan blockreader.GetFromDiskResp)

	connsMain.OutGetFromDiskReq = localReadDest
	connsMain.OutGetFromDiskResp = readResp

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
		Sends:       netSends,
		NoConnResp:  readResp,
		NodeConnMap: nodeConnMapper,
	}
	go rbr.Start(ctx)

	lbrr := blockreader.LocalBlockReadResponses{
		InDisks:            addedDisksReader,
		LocalReadResponses: readResp,
		Sends:              netSends,
		NodeConnMap:        nodeConnMapper,
		NodeId:             nodeId,
	}
	go lbrr.Start(ctx)

	inFileBroadcasts := make(chan webdav.FileBroadcast, 1)
	connsMain.OutFileBroadcasts = inFileBroadcasts
	_ = webdav.New(
		nodeId,
		webdavGetReq,
		webdavPutReq,
		webdavGetResp,
		webdavPutResp,
		netSends,
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
