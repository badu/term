package geom

// Node is used as items in a Tree.
type Node struct {
	Children  []Node
	Rectangle Rectangle
}

// LeveledList is a list, which contains multiple Item.
type LeveledList []Item

// Item combines a text with a specific level.
// The level is the indent, which would normally be seen in a BulletList.
type Item struct {
	Level     int
	Rectangle Rectangle
}

// Tree is able to render a list.
type Tree struct {
	Root *Node
	// TODO : keep pixels here
}

// WithRoot returns a new list with a specific Root.
func (p *Tree) WithRoot(root *Node) *Tree {
	p.Root = root
	return p
}

// traverse the list.
// TODO : make two traversals 1. adjust sizes if needed 2. allocate pixels
func (p *Tree) traverse() {
	walkOverTree(p.Root.Children, p)
}

// walkOverTree is a recursive function, which analyzes a Tree and connects the items with specific characters.
func walkOverTree(list []Node, p *Tree) {
	for i, item := range list {
		if len(list) > i+1 { // if not last in list
			if len(item.Children) == 0 { // if there are no children

			} else { // if there are children

				walkOverTree(item.Children, p)
			}
		} else if len(list) == i+1 { // if last in list
			if len(item.Children) == 0 { // if there are no children

			} else { // if there are children

				walkOverTree(item.Children, p)
			}
		}
	}
}

// NewTreeFromLeveledList converts a TreeItems list to a Node and returns it.
func NewTreeFromLeveledList(items LeveledList) Node {
	if len(items) == 0 {
		return Node{}
	}

	root := &Node{
		Children:  []Node{},
		Rectangle: items[0].Rectangle,
	}

	for i, record := range items {
		last := root

		if record.Level < 0 {
			record.Level = 0
			items[i].Level = 0
		}

		if len(items)-1 != i {
			if items[i+1].Level-1 > record.Level {
				items[i+1].Level = record.Level + 1
			}
		}

		for i := 0; i < record.Level; i++ {
			lastIndex := len(last.Children) - 1
			last = &last.Children[lastIndex]
		}
		last.Children = append(last.Children, Node{
			Children:  []Node{},
			Rectangle: record.Rectangle,
		})
	}

	return *root
}
