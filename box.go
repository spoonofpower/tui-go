package tui

import (
	"image"
	"math"
)

var _ Widget = &Box{}

// Alignment is used to set the direction in which widgets are laid out.
type Alignment int

// Available alignment options.
const (
	Horizontal Alignment = iota
	Vertical
)

// Box is a layout for placing widgets.
type Box struct {
	WidgetBase

	children []Widget

	border bool
	title  string

	alignment Alignment
}

// NewVBox returns a new vertical Box.
func NewVBox(c ...Widget) *Box {
	return &Box{
		children:  c,
		alignment: Vertical,
	}
}

// NewHBox returns a new horizontal Box.
func NewHBox(c ...Widget) *Box {
	return &Box{
		children:  c,
		alignment: Horizontal,
	}
}

// Append adds a new widget to the layout.
func (b *Box) Append(w Widget) {
	b.children = append(b.children, w)
}

func (b *Box) Clear() {
	b.children = nil
}

// SetBorder sets whether the border is visible or not.
func (b *Box) SetBorder(enabled bool) {
	b.border = enabled
}

func (b *Box) SetTitle(title string) {
	b.title = title
}

// Alignment returns the currently set alignment or the Box.
func (b *Box) Alignment() Alignment {
	return b.alignment
}

// Draw recursively draws the children it contains.
func (b *Box) Draw(p *Painter) {
	sz := b.Size()

	if b.border {
		p.DrawRect(0, 0, sz.X, sz.Y)
		p.WithMask(image.Rect(2, 0, sz.X-3, 0)).DrawText(2, 0, b.title)

		p.Translate(1, 1)
		defer p.Restore()
	}

	var off image.Point
	for _, child := range b.children {
		switch b.Alignment() {
		case Horizontal:
			p.Translate(off.X, 0)
		case Vertical:
			p.Translate(0, off.Y)
		}

		child.Draw(p.WithMask(image.Rectangle{
			Min: image.ZP,
			Max: child.Size().Sub(image.Point{1, 1}),
		}))

		p.Restore()

		off = off.Add(child.Size())
	}
}

// MinSizeHint returns the minimum size for the layout.
func (b *Box) MinSizeHint() image.Point {
	var minSize image.Point

	for _, child := range b.children {
		size := child.MinSizeHint()
		if b.Alignment() == Horizontal {
			minSize.X += size.X
			if size.Y > minSize.Y {
				minSize.Y = size.Y
			}
		} else {
			minSize.Y += size.Y
			if size.X > minSize.X {
				minSize.X = size.X
			}
		}
	}

	if b.border {
		minSize = minSize.Add(image.Point{2, 2})
	}

	return minSize
}

// SizeHint returns the recommended size for the layout.
func (b *Box) SizeHint() image.Point {
	var sizeHint image.Point

	for _, child := range b.children {
		size := child.SizeHint()
		if b.Alignment() == Horizontal {
			sizeHint.X += size.X
			if size.Y > sizeHint.Y {
				sizeHint.Y = size.Y
			}
		} else {
			sizeHint.Y += size.Y
			if size.X > sizeHint.X {
				sizeHint.X = size.X
			}
		}
	}

	if b.border {
		sizeHint = sizeHint.Add(image.Point{2, 2})
	}

	return sizeHint
}

// OnEvent handles an event and propagates it to all children.
func (b *Box) OnEvent(ev Event) {
	for _, child := range b.children {
		child.OnEvent(ev)
	}
}

// Resize updates the size of the layout.
func (b *Box) Resize(size image.Point) {
	b.size = size
	inner := b.size
	if b.border {
		inner = b.size.Sub(image.Point{2, 2})
	}
	b.layoutChildren(inner)
}

func (b *Box) layoutChildren(size image.Point) {
	space := doLayout(b.children, dim(b.Alignment(), size), b.Alignment())

	for i, s := range space {
		switch b.Alignment() {
		case Horizontal:
			b.children[i].Resize(image.Point{s, size.Y})
		case Vertical:
			b.children[i].Resize(image.Point{size.X, s})
		}
	}
}

func doLayout(ws []Widget, space int, a Alignment) []int {
	sizes := make([]int, len(ws))

	if len(sizes) == 0 {
		return sizes
	}

	remaining := space

	// Distribute MinSizeHint
	for {
		var changed bool
		for i, sz := range sizes {
			if sz < dim(a, ws[i].MinSizeHint()) {
				sizes[i] = sz + 1
				remaining--
				if remaining <= 0 {
					goto Resize
				}
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Distribute Minimum
	for {
		var changed bool
		for i, sz := range sizes {
			p := alignedSizePolicy(a, ws[i])
			if p == Minimum && sz < dim(a, ws[i].SizeHint()) {
				sizes[i] = sz + 1
				remaining--
				if remaining <= 0 {
					goto Resize
				}
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Distribute Preferred
	for {
		var changed bool
		for i, sz := range sizes {
			p := alignedSizePolicy(a, ws[i])
			if (p == Preferred || p == Maximum) && sz < dim(a, ws[i].SizeHint()) {
				sizes[i] = sz + 1
				remaining--
				if remaining <= 0 {
					goto Resize
				}
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Distribute Expanding
	for {
		var changed bool
		for i, sz := range sizes {
			p := alignedSizePolicy(a, ws[i])
			if p == Expanding {
				sizes[i] = sz + 1
				remaining--
				if remaining <= 0 {
					goto Resize
				}
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// Distribute remaining space
	for {
		min := math.MaxInt8
		for i, s := range sizes {
			p := alignedSizePolicy(a, ws[i])
			if (p == Preferred || p == Minimum) && s <= min {
				min = s
			}
		}
		var changed bool
		for i, sz := range sizes {
			if sz != min {
				continue
			}
			p := alignedSizePolicy(a, ws[i])
			if p == Preferred || p == Minimum {
				sizes[i] = sz + 1
				remaining--
				if remaining <= 0 {
					goto Resize
				}
				changed = true
			}
		}
		if !changed {
			break
		}
	}

Resize:

	return sizes
}

func dim(a Alignment, pt image.Point) int {
	if a == Horizontal {
		return pt.X
	}
	return pt.Y
}

func alignedSizePolicy(a Alignment, w Widget) SizePolicy {
	hpol, vpol := w.SizePolicy()
	if a == Horizontal {
		return hpol
	}
	return vpol
}
