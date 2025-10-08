## Rebalancer

1. Rebalancer starts with a `BalanceReqId`
2. For each `BlockId` in the filesystem find the destinations
3. Rebalancer sends an `ExistsReq` message to `ExistsSender`

> `ExistsReq`
> - Caller
> - BalanceReqId
> - ExistsId
> - DestNodeId
> - DestDiskId
> - DestBlockId

4. And wait for an `ExistsResp` in the `Rebalancer`

> `ExistsResp`
> - ExistsReq
> - Ok
> - Msg

5. In order to wait for the response save with a two dimensional map, first key is `BalanceReqId`, second key is `BlockId`, value is a set of `ExistsReq`. In we also need to save a `toDelete` map that has the same keys but the value is a set of `BlockId` + `DestNodeId` + `DestDiskId`. Those are the dests that don't need the block id anymore once the ones that need it have it
6. When we get a `ExistsResp` where `Exists == true` then remove that item from the set.
7. If the set is empty send a `BalanceReqId` + `BlockId`  to the `DeleteSender`
8. If the `ExistsResp` response has `Exists == false` then send a `StoreItReq` to `StoreItSender`

> `StoreItReq`
> - Caller
> - BalanceReqId
> - StoreItId
> - DestNodeId
> - DestDiskId
> - DestBlockId

9. In the network the `StoreItReq` will turn into a `StoreItCmd` which will turn into a `StoreItResp`

> `StoreItResp`
> - StoreItReq
> - Ok
> - Msg

10. When `Rebalancer` gets a `StoreItResp` that is OK then delete the associated `ExistsReq`