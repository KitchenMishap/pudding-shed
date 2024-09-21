package wordfile

import (
	"github.com/KitchenMishap/pudding-shed/memfile"
)

type MemShadowedWordFile struct {
	underlying *WordFile
	shadow     []int64
}

func NewMemShadowedWordFile(file memfile.AppendableLookupFile, wordSize int64, wordCount int64) (*MemShadowedWordFile, error) {
	p := new(MemShadowedWordFile)
	p.underlying = NewWordFile(file, wordSize, wordCount)
	p.shadow = make([]int64, wordCount)
	for i := int64(0); i < wordCount; i++ {
		var err error
		p.shadow[i], err = p.underlying.ReadWordAt(i)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (wf *MemShadowedWordFile) ReadWordAt(off int64) (int64, error) {
	// Just read the shadow
	return wf.shadow[off], nil
}

func (wf *MemShadowedWordFile) WriteWordAt(val int64, off int64) error {
	// Write to shadow and underlying
	length := int64(len(wf.shadow))
	if off >= length {
		newVals := off + 1 - length
		for i := int64(0); i < newVals; i++ {
			wf.shadow = append(wf.shadow, 0)
		}
	}
	wf.shadow[off] = val

	err := wf.underlying.WriteWordAt(val, off)
	return err
}

func (wf *MemShadowedWordFile) CountWords() (words int64, err error) {
	return wf.underlying.CountWords()
}

func (wf *MemShadowedWordFile) Close() error {
	return wf.underlying.Close()
}

func (wf *MemShadowedWordFile) Sync() error {
	return wf.underlying.Sync()
}

func (wf *MemShadowedWordFile) WordSize() int64 {
	return wf.underlying.WordSize()
}
