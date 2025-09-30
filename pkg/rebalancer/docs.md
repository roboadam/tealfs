## Rebalancer

1. Rebalancer starts with a `BalanceReqId`
2. For each `BlockId` in the filesystem find the destinations
3. Rebalancer sends an `ExistsReq` message to the destination

> `ExistsReq`
> - ExistsId
> - BalanceReqId
> - DestNodeId
> - DestDiskId
> - BlockId

4. And wait for an `ExistsResp`

> `ExistsResp`
> - ExistsReq
> - Exists (bool)

5. In order to wait for the response save with a two dimensional map, first key is `BalanceReqId`, second key is `BlockId`, value is a set of `ExistsReq`
6. When we get a `ExistsResp` where `Exists == true` then remove that item from the set.
7. If the set is empty send a `StoreItReq` to the `StoreItHandler`

> `StoreItReq`
> - StoreItId
> - BalanceReqId
> - DestNodeId
> - DestDiskId
> - BlockId

8. If the `ExistsResp

