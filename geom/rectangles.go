package geom

import (
	"context"
	"log"

	"github.com/badu/term"
	"github.com/badu/term/style"
)

// Node is used as items in a Tree.
type Node struct {
	Children  []Node     //
	Rectangle *Rectangle //
	level     int
}

// LeveledList is a list, which contains multiple Item.
type LeveledList []Item

// Item combines a text with a specific level.
// The level is the indent, which would normally be seen in a BulletList.
type Item struct {
	Level     int        //
	Rectangle *Rectangle //
}

// Tree is able to render a list.
type Tree struct {
	Root      *Node //
	currLevel int
}

// newTree
func NewTree(ctx context.Context, cols, rows int, ch chan term.Position, oriented style.Orientation, core term.ResizeDispatcher) Tree {
	var opts []RectangleOption
	opts = append(opts, WithBottomCorner(cols, rows))
	opts = append(opts, WithTopCorner(0, 0))
	opts = append(opts, WithAcquisitionChan(ch))
	opts = append(opts, WithCore(core))
	if oriented != style.NoOrientation {
		opts = append(opts, WithOrientation(oriented))
	}
	r, _ := NewRectangle(ctx, opts...)
	result := Tree{Root: &Node{Rectangle: r, level: 1}}
	return result
}

// Register
func (t *Tree) Register(targets ...*Rectangle) {
	if t.currLevel == 0 {
		log.Printf("registering %d children to root.", len(targets))
		t.Root.Rectangle.SetChildren(targets...)
		t.currLevel++
	} else {
		log.Printf("at level %d, setting %d children", t.currLevel, len(targets))
		findNode(t.Root.Children, t, targets)
	}
}

// findNode is a recursive function, which analyzes a Tree and connects the items with specific characters.
func findNode(list []Node, p *Tree, children []*Rectangle) {
	for i, item := range list {
		if item.level == p.currLevel {
			log.Printf("deep child found")
			// item found
			p.currLevel++
			item.Rectangle.SetChildren(children...)
			break
		}
		if len(list) > i+1 { // if not last in list
			if len(item.Children) != 0 { // if there are children
				findNode(item.Children, p, children)
			}
		} else if len(list) == i+1 { // if last in list
			if len(item.Children) != 0 { // if there are children
				findNode(item.Children, p, children)
			}
		}
	}
}

// traverse the list.
// TODO : make two traversals 1. adjust sizes if needed 2. allocate pixels
func (t *Tree) traverse() {
	walkOverTree(t.Root.Children, t)
}

// walkOverTree is a recursive function, which analyzes a Tree.
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
