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

package conns

import "tealfs/pkg/model"

func (c *Conns) handleIam(iam *model.IAm) {
	c.NodeConnMapper.SetNodeAddress(iam.Node.NodeId, iam.Node.Address)
	for _, sibling := range iam.Siblings {
		if !c.NodeConnMapper.KnownNode(sibling.NodeId) {
			c.NodeConnMapper.SetNodeAddress(sibling.NodeId, sibling.Address)
		}
	}
	go c.connectToUnConnected()
}

func (c *Conns) connectToUnConnected() {
	addresses := c.NodeConnMapper.AddressesWithoutConnections()
	for _, address := range addresses.GetValues() {
		c.inConnectTo <- model.ConnectToNodeReq{Address: address}
	}
}
