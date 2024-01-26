package handler

type File struct {
	svc Service
}

func NewFile1(s Service) *File { return &File{s} }

func (f *File) WriteToFile() error {
	return f.svc.WriteToFile()
}

func (f *File) RestoreFromFile() error {
	return f.svc.RestoreFromFile()
}
