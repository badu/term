package core

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"
)

var (
	MissingHandlers = errors.New("at least one handler is required")
)

// IntChangeListener is the interface that a listener must implement
type IntChangeListener interface {
	OnChange(value int)
}

// IntChannel is implemented by intImpl
type IntChannel interface {
	Ch() chan int
}

// IntGetter is implemented by intImpl
type IntGetter interface {
	Get() int
}

// IntSetter is accessed by controller
type IntSetter interface {
	Set(int)
}

// Int is interface used by propagate context mixer
type Int interface {
	IntChannel
	IntGetter
	IntSetter
}

// intImpl is a model which propagate change events
type intImpl struct {
	ch   chan int
	curr int
}

// the setter, which will write into the channel that listeners get
func (i *intImpl) Set(value int) {
	if i.curr == value {
		return
	}
	i.curr = value
	i.ch <- i.curr
}

// IntGetter implementation
func (i *intImpl) Get() int {
	return i.curr
}

// IntChannel implementation
func (i *intImpl) Ch() chan int {
	return i.ch
}

// Int constructor
func NewInt(initial int) Int {
	return &intImpl{curr: initial, ch: make(chan int)}
}

// mixes a context, an IntChannel (interface) and the listeners
// It is caller responsibility to run the returned function as a goroutine
func PropagateIntChange(ctx context.Context, int Int, handlers ...IntChangeListener) (func(), error) {
	if len(handlers) == 0 {
		return nil, MissingHandlers
	}
	res := func() {
		// propagate initial value, immediately on register ("init" like)
		for _, h := range handlers {
			h.OnChange(int.Get())
		}
		// main loop, which is waiting for cancellation or propagate integer value changes
		for {
			select {
			case <-ctx.Done():
				return
			case value := <-int.Ch():
				for _, h := range handlers {
					h.OnChange(value)
				}
			}
		}
	}
	return res, nil
}

// RuneChangeListener is the interface that a listener must implement
type RuneChangeListener interface {
	OnChange(value rune)
}

// RuneChannel is implemented by runeImpl
type RuneChannel interface {
	Ch() chan rune
}

// RuneGetter is implemented by runeImpl
type RuneGetter interface {
	Get() rune
}

// RuneSetter is accessed by controller
type RuneSetter interface {
	Set(rune)
}

// Char is interface used by propagate context mixer
type Char interface {
	RuneChannel
	RuneGetter
	RuneSetter
}

// runeImpl is a model which propagate change events
type runeImpl struct {
	ch   chan rune
	curr rune
}

// the setter, which will write into the channel that listeners get
func (i *runeImpl) Set(value rune) {
	if i.curr == value {
		return
	}
	i.curr = value
	i.ch <- i.curr
}

// RuneGetter implementation
func (i *runeImpl) Get() rune {
	return i.curr
}

// RuneChannel implementation
func (i *runeImpl) Ch() chan rune {
	return i.ch
}

// Char constructor
func NewChar(initial rune) Char {
	return &runeImpl{curr: initial, ch: make(chan rune)}
}

// mixes a context, an RuneChannel (interface) and the listeners
// It is caller responsibility to run the returned function as a goroutine
func PropagateRuneChange(ctx context.Context, int Char, handlers ...RuneChangeListener) (func(), error) {
	if len(handlers) == 0 {
		return nil, MissingHandlers
	}
	res := func() {
		// propagate initial value, immediately on register ("init" like)
		for _, h := range handlers {
			h.OnChange(int.Get())
		}
		// main loop, which is waiting for cancellation or propagate integer value changes
		for {
			select {
			case <-ctx.Done():
				return
			case value := <-int.Ch():
				for _, h := range handlers {
					h.OnChange(value)
				}
			}
		}
	}
	return res, nil
}

// CharsChangeListener is the interface that a listener must implement
type CharsChangeListener interface {
	OnChange(value []rune)
}

// CharsChannel is implemented by charsImpl
type CharsChannel interface {
	Ch() chan []rune
}

// CharsGetter is implemented by charsImpl
type CharsGetter interface {
	Get() []rune
}

// CharsSetter is required by controller
type CharsSetter interface {
	Set([]rune) error
}

// CharsSize is implemented by charsImpl
type CharsSize interface {
	Size() []rune
}

// Chars is interface used by propagate context mixer
type Chars interface {
	CharsChannel
	CharsGetter
	CharsSetter
}

// charsImpl is a model which propagate change events
type charsImpl struct {
	ch   chan []rune
	curr []rune
	size int
}

// the setter, which will write into the channel that listeners get
func (i *charsImpl) Set(value []rune) error {
	// validate and compare
	equal := len(i.curr) == len(value)
	newSize := 0
	for idx, r := range value {
		if !utf8.ValidRune(r) {
			return fmt.Errorf("invalid rune provided : %#v", r)
		}
		// only if equal len, we're comparing rune by rune
		if equal && len(i.curr) > idx && i.curr[idx] != r {
			// never enter here again, just continue validating runes
			equal = false
		}
		newSize += utf8.RuneLen(r)
	}
	if equal {
		return errors.New("no change")
	}
	// update value and size
	i.size = newSize
	i.curr = value
	i.ch <- i.curr
	return nil
}

// CharsSize implementation
func (i *charsImpl) Size() int {
	return i.size
}

// CharsGetter implementation
func (i *charsImpl) Get() []rune {
	return i.curr
}

// CharsChannel implementation
func (i *charsImpl) Ch() chan []rune {
	return i.ch
}

// NewChars constructor
func NewChars(initial []rune) (Chars, error) {
	iSize := 0
	for _, r := range initial {
		if !utf8.ValidRune(r) {
			return nil, errors.New("invalid rune provided")
		}
		iSize += utf8.RuneLen(r)
	}
	res := charsImpl{curr: initial, ch: make(chan []rune), size: iSize}
	return &res, nil
}

// mixes a context, an CharsChannel (interface) and the listeners
// It is caller responsibility to run the returned function as a goroutine
// It is handler responsibility to validate for nil rune slices
func PropagateCharsChange(ctx context.Context, int Chars, handlers ...CharsChangeListener) (func(), error) {
	if len(handlers) == 0 {
		return nil, MissingHandlers
	}
	res := func() {
		// propagate initial value, immediately on register ("init" like)
		for _, h := range handlers {
			h.OnChange(int.Get())
		}
		// main loop, which is waiting for cancellation or propagate integer value changes
		for {
			select {
			case <-ctx.Done():
				return
			case value := <-int.Ch():
				for _, h := range handlers {
					h.OnChange(value)
				}
			}
		}
	}
	return res, nil
}
