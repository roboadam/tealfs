package proto

type SaveData struct {
	Data []byte
}

func (s *SaveData) ToBytes() []byte {
	return AddType(SaveDataType, s.Data)
}

func ToSaveData(data []byte) *SaveData {
	return &SaveData{
		Data: data,
	}
}
