package proto

import "tealfs/pkg/store"

type ReadData struct {
	StoreId store.Id
}

func (r *ReadData) ToBytes() []byte {
	// Todo: This needs to be enhanced to serialize all the Block, not just the data
	// The logic should probably live in the block and be called from here
	data := IntToBytes(uint32(r.StoreId))
	return AddType(ReadDataType, s.Block.Data)
}

func ToSaveData(data []byte) *SaveData {
	// Todo: This needs to be enhanced to deserialize all the Block, not just the data
	// The logic should probably live in the block and be called from here
	return &SaveData{
		Block: store.Block{Data: data},
	}
}
