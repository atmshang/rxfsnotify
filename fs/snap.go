package fs

import (
	"sync"
)

type Snapshot struct {
	oldSnapshot *FileSystem
	curSnapshot *FileSystem
	rwLocker    sync.RWMutex
}

func (fss *Snapshot) Init(rootDirPath string) error {

	fss.rwLocker.Lock()
	defer fss.rwLocker.Unlock()

	tempFs, err := NewFileSystem(rootDirPath)
	if err != nil {
		return err
	}
	fss.oldSnapshot = tempFs
	fss.curSnapshot = fss.oldSnapshot.DeepCopy()
	return nil
}

func (fss *Snapshot) UpdateChangedDir(changedDirPath string) error {
	fss.rwLocker.Lock()
	defer fss.rwLocker.Unlock()

	err := fss.curSnapshot.Update(changedDirPath)
	if err != nil {
		return err
	}
	return nil
}

func (fss *Snapshot) DiffAndSync() []Diff {
	fss.rwLocker.Lock()
	defer fss.rwLocker.Unlock()

	diffs := fss.oldSnapshot.Diff(fss.curSnapshot)
	if len(diffs) > 0 {
		fss.oldSnapshot = fss.curSnapshot.DeepCopy()
	}
	return diffs
}
