package fs

import (
	"fmt"
)

type Diff struct {
	AbsPath string
	Path    string
	Op      int
}

func diffNodes(oldNode *Node, newNode *Node, path string) []Diff {
	var diffs []Diff

	// If both nodes are null, there's no difference.
	if oldNode == nil && newNode == nil {
		return diffs
	}

	// If one of the nodes is null, there's a difference.
	if oldNode == nil || newNode == nil {
		op := 1
		absPath := ""
		if oldNode != nil {
			op = 0
			absPath = oldNode.AbsPath
		} else {
			absPath = newNode.AbsPath
		}
		diffs = append(diffs, Diff{
			AbsPath: absPath,
			Path:    path,
			Op:      op,
		})
		return diffs
	}

	// If both nodes exist, check their attributes.
	if oldNode.IsFile != newNode.IsFile {
		diffs = append(diffs, Diff{
			AbsPath: newNode.AbsPath,
			Path:    path,
			Op:      2,
		})
		return diffs
	}

	// If both nodes exist and are directories, check their children.
	oldChildren := make(map[string]*Node)
	for _, child := range oldNode.Children {
		oldChildren[child.Name] = child
	}

	newChildren := make(map[string]*Node)
	for _, child := range newNode.Children {
		newChildren[child.Name] = child
	}

	for name, oldChild := range oldChildren {
		newChild, ok := newChildren[name]
		if ok {
			childPath := path
			if path != "" {
				childPath += "/"
			}
			childPath += name
			childDiffs := diffNodes(oldChild, newChild, childPath)
			diffs = append(diffs, childDiffs...)
		} else {
			diffs = append(diffs, Diff{
				AbsPath: oldChild.AbsPath,
				Path:    path + "/" + name,
				Op:      0,
			})
		}
	}

	for name, newChild := range newChildren {
		_, ok := oldChildren[name]
		if !ok {
			diffs = append(diffs, Diff{
				AbsPath: newChild.AbsPath,
				Path:    path + "/" + name,
				Op:      1,
			})
		}
	}

	return diffs
}

func (fs *FileSystem) Diff(fs2 *FileSystem) []Diff {
	return diffNodes(fs.Root, fs2.Root, "")
}

func (d Diff) String() string {
	opString := ""
	switch d.Op {
	case 0:
		opString = "file/directory deleted"
	case 1:
		opString = "new file/directory"
	case 2:
		opString = "file modified"
	}
	return fmt.Sprintf("AbsPath: %s, Path: %s, Operation: %s", d.AbsPath, d.Path, opString)
}
