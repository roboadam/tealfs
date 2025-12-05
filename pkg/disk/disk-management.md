# Disk Management

Lets say we have `node1` and `node2` initially with no disks, we will add one disk to `node1` using the first node's UI, then restart `node2`.

1. From the UI a user adds a disk to `node1`, this triggers a message `AddDiskMsg` with a new `DiskId`.

```
AddDiskMsg
  NodeId
  DiskId
  Path
```

2. This messages goes to a `MessageSenderSvc` that routes it to the local `DiskManagerSvc`

3. The local `DiskManagerSvc` has two sets one is called `LocalDiskSvcList` which represents the local disk services that actually write to the disk, the other is called `DiskInfoList` which is a set POGOs that represent all disks in the cluster. It receives the `AddDiskMsg`, and if the `DiskId` doesn't already exist in each list, add it.

```
DiskInfo
  NodeId
  DiskId
  Path
```

4. If we for real added a disk to LocalDiskServices then send a message `DiskAddedMessage` to all other nodes.

```
DiskAddedMessage
  NodeId
  DiskId
  Path
```

5. Each time something is added to `DiskManagerSvc.DiskInfoList` persist the list to `disks.json`

6. On startup read `disks.json` and load the two lists inside `DiskManagerSvc`. Block on startup so that the subsequent `Iam` message has the info it needs.

7. `Iam` contains a `DiskAddedList`
