## Rebalancer

1. Rebalance starts with a `BalanceReqId`
2. Rebalancer sends a `BlockId` + `BalanceReqId` message to `ExistsSender`
3. `ExistsSender` finds the two intended destinations for `BlockId`
4. For each destination send an `ExistsReq`

> `ExistsReq`
> - BlockId
> - DestNodeId
> - DestDiskId
> - ExistsId
> - BalanceReqId

5. And wait for an `ExistsResp`

> `ExistsResp`
> - ExistsReq
> - Exists (bool)

6. In order to wait for the response save with a two dimensional map, first key is `BalanceReqId`, second key is `BlockId`, value is a set of `ExistsReq`
7. When we get a response where `Exists == true` then remove that

