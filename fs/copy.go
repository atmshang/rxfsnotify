package fs

func (n *Node) deepCopy() *Node {
	newNode := &Node{
		Name:     n.Name,
		AbsPath:  n.AbsPath,
		IsFile:   n.IsFile,
		Size:     n.Size,
		ModTime:  n.ModTime,
		Children: make([]*Node, len(n.Children)),
	}

	for i, child := range n.Children {
		newNode.Children[i] = child.deepCopy()
	}

	return newNode
}

func (fs *FileSystem) DeepCopy() *FileSystem {
	return &FileSystem{
		Root: fs.Root.deepCopy(),
	}
}
