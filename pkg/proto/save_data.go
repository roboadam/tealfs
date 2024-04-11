package proto

import "tealfs/pkg/store"

type SaveData struct {
	Id       store.Id
	Data     []byte
	Children []store.Id
}

func (s *SaveData) ToBytes() []byte {
	return AddType(SaveDataType, s.Data)
}

func ToSaveData(data []byte) *SaveData {
	return &SaveData{
		Data: data,
	}
}
