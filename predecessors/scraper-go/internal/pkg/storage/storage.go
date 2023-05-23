package storage

type FileStorage interface {
	WriteData(data interface{}, filePath string) error
	ReadData(filePath string) (interface{}, error)
	GetLatestFiles(folderPrefix, textIn string) ([]string, error)
	GetLatestFile(folderPrefix, textIn string) (string, error)
}
