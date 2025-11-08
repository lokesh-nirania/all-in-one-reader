package reader

import "io"

type ProgressReader struct {
	Reader    io.Reader
	TotalSize int64
	ReadSize  int64
	Notify    func(n int64, totalSize int64)
}

func (p *ProgressReader) Read(buf []byte) (int, error) {
	n, err := p.Reader.Read(buf)
	if n > 0 {
		p.ReadSize += int64(n)
		if p.Notify != nil {
			p.Notify(p.ReadSize, p.TotalSize)
		}
	}
	return n, err
}
