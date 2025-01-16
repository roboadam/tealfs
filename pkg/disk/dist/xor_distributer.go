// Copyright (C) 2024 Adam Hess
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

package dist

import (
	"encoding/binary"
	"errors"
	"hash"
	"hash/crc32"
	"tealfs/pkg/model"
)

type XorDistributer struct {
	weights  map[model.NodeId]int
	checksum hash.Hash32
}

func NewXorDistributer() XorDistributer {
	return XorDistributer{
		weights:  make(map[model.NodeId]int),
		checksum: crc32.NewIEEE(),
	}
}

func (d *XorDistributer) KeysForIds(id1 model.BlockKeyId, id2 model.BlockKeyId) (model.BlockKey, model.BlockKey, error) {
	node1, node2, parity, err := d.generateNodeIds(id1, id2)
	if err != nil {
		return model.BlockKey{}, model.BlockKey{}, err
	}

	parityPointer := model.DiskPointer{
		NodeId:   parity,
		FileName: string(id1) + "." + string(id2),
	}

	key1 := model.BlockKey{
		Id:   id1,
		Type: model.XORed,
		Data: []model.DiskPointer{model.DiskPointer{
			NodeId:   node1,
			FileName: string(id1),
		}},
		Parity: parityPointer,
	}

	key2 := model.BlockKey{
		Id:   id2,
		Type: model.XORed,
		Data: []model.DiskPointer{model.DiskPointer{
			NodeId:   node2,
			FileName: string(id2),
		}},
		Parity: parityPointer,
	}

	return key1, key2, nil
}

func (d *XorDistributer) generateNodeIds(id1 model.BlockKeyId, id2 model.BlockKeyId) (node1 model.NodeId, node2 model.NodeId, parity model.NodeId, err error) {
	if len(d.weights) < 3 {
		return "", "", "", errors.New("not enough nodes to generate parity")
	}

	idb := []byte(id1)
	checksum := d.Checksum(idb)
	intHash := int(binary.BigEndian.Uint32(checksum))

	node1 = d.nodeIdForHashAndWeights(intHash, d.weights)

	if len(d.weights) == 1 {
		for key := range d.weights {
			node2 = key
			parity = key
			err = nil
			return
		}
	}

	weights2 := d.weights
	delete(weights2, node1)

	idb = []byte(id2)
	checksum = d.Checksum(idb)
	intHash = int(binary.BigEndian.Uint32(checksum))

	node2 = d.nodeIdForHashAndWeights(intHash, d.weights)

	if len(d.weights) == 1 {
		for key := range d.weights {
			parity = key
			err = nil
			return
		}
	}

	idb = []byte(id1 + id2)
	checksum = d.Checksum(idb)
	intHash = int(binary.BigEndian.Uint32(checksum))

	parity = d.nodeIdForHashAndWeights(intHash, d.weights)
	err = nil
	return
}

func (d *XorDistributer) nodeIdForHashAndWeights(hash int, weights map[model.NodeId]int) model.NodeId {
	total := totalWeight(weights)
	randomNum := hash % total

	cumulativeWeight := 0
	keys := sortedKeys(weights)
	for _, key := range keys {
		cumulativeWeight += weights[key]
		if randomNum < cumulativeWeight {
			return key
		}
	}
	panic("should never get here")
}

func (d *XorDistributer) Checksum(data []byte) []byte {
	d.checksum.Reset()
	d.checksum.Write(data)
	return d.checksum.Sum(nil)
}

func (d *XorDistributer) SetWeight(id model.NodeId, weight int) {
	d.weights[id] = weight
}
