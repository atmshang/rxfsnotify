package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Node struct {
	Name     string
	AbsPath  string
	Children []*Node
	IsFile   bool
	Size     int64
	ModTime  time.Time
}

type FileSystem struct {
	Root *Node
}

func newNode(name string, absPath string, isFile bool) *Node {
	return &Node{Name: name, AbsPath: absPath, IsFile: isFile}
}

func (n *Node) addChild(name string, absPath string, isFile bool) *Node {
	newNode := newNode(name, absPath, isFile)
	n.Children = append(n.Children, newNode)
	return newNode
}

func NewFileSystem(rootAbsPath string) (*FileSystem, error) {

	instance := FileSystem{
		Root: newNode(string(os.PathSeparator), rootAbsPath, false),
	}
	err := instance.build(rootAbsPath)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (fs *FileSystem) Update(innerAbsPath string) error {
	if !strings.HasPrefix(innerAbsPath, fs.Root.AbsPath) {
		return fmt.Errorf("path %s is not a subpath of the root path %s", innerAbsPath, fs.Root.AbsPath)
	}

	innerPath, err := filepath.Rel(fs.Root.AbsPath, innerAbsPath)
	if err != nil {
		return err
	}

	// Remove old node.
	parentNode := fs.Root
	parts := strings.Split(innerPath, string(os.PathSeparator))
	for i, part := range parts {
		if i == len(parts)-1 {
			// Remove the last part from its parent node.
			for j, child := range parentNode.Children {
				if child.Name == part {
					parentNode.Children = append(parentNode.Children[:j], parentNode.Children[j+1:]...)
					break
				}
			}
		} else {
			// Find the next parent node.
			found := false
			for _, child := range parentNode.Children {
				if child.Name == part {
					parentNode = child
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("part %s not found in path %s", part, innerPath)
			}
		}
	}

	// Build new node.
	return filepath.Walk(innerAbsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		isFile := !info.IsDir()
		size := info.Size()
		modTime := info.ModTime()

		relPath, err := filepath.Rel(fs.Root.AbsPath, path)
		if err != nil {
			return err
		}

		fs.add(relPath, isFile, size, modTime)

		return nil
	})
}

func (fs *FileSystem) build(rootPath string) error {
	fs.Root.AbsPath = rootPath
	return filepath.Walk(fs.Root.AbsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		isFile := !info.IsDir()
		size := info.Size()
		modTime := info.ModTime()

		relPath, err := filepath.Rel(fs.Root.AbsPath, path)
		if err != nil {
			return err
		}

		fs.add(relPath, isFile, size, modTime)

		return nil
	})
}

func (fs *FileSystem) add(path string, isFile bool, size int64, modTime time.Time) {
	parts := strings.Split(path, string(os.PathSeparator))
	currentNode := fs.Root

	for _, part := range parts {
		found := false
		for _, child := range currentNode.Children {
			if child.Name == part {
				currentNode = child
				found = true
				break
			}
		}

		if !found {
			absPath := filepath.Join(currentNode.AbsPath, part)
			currentNode = currentNode.addChild(part, absPath, isFile)
			currentNode.Size = size
			currentNode.ModTime = modTime
		}
	}
}

func (n *Node) Print(prefix string) {
	var nodeType string
	if n.IsFile {
		nodeType = "File"
	} else {
		nodeType = "Directory"
	}
	fmt.Printf("%s%s (%s)(%d)(%s)(%s)\n", prefix, n.Name, nodeType, n.Size, n.ModTime.Format("2006-01-02 15:04:05"), n.AbsPath)
	for _, child := range n.Children {
		child.Print(prefix + "  ")
	}
}

func (fs *FileSystem) Print() {
	fs.Root.Print("")
}
