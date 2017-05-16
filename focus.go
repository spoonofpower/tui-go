package tui

type FocusChain interface {
	FocusNext() Widget
	FocusPrev() Widget
	FocusDefault() Widget
}

type KbFocusController struct {
	chain FocusChain
}

func (c *KbFocusController) OnEvent(e Event) {
	if e.Type != EventKey || c.chain == nil {
		return
	}
	switch e.Key {
	case KeyTab:
		c.chain.FocusNext()
	case KeyBacktab:
		c.chain.FocusPrev()
	}
}

var DefaultFocusChain = &SimpleFocusChain{
	widgets: make([]Widget, 0),
}

type SimpleFocusChain struct {
	widgets []Widget
}

func (c *SimpleFocusChain) Set(ws ...Widget) {
	c.widgets = ws
}

func (c *SimpleFocusChain) FocusNext() Widget {
	for i, w := range c.widgets {
		if !w.IsFocused() {
			continue
		}
		w.SetFocused(false)
		if i < len(c.widgets)-1 {
			c.widgets[i+1].SetFocused(true)
			return c.widgets[i+1]
		}
		c.widgets[0].SetFocused(true)
		return c.widgets[0]
	}
	return nil
}

func (c *SimpleFocusChain) FocusPrev() Widget {
	for i, w := range c.widgets {
		if !w.IsFocused() {
			continue
		}
		w.SetFocused(false)
		if i <= 0 {
			c.widgets[len(c.widgets)-1].SetFocused(true)
			return c.widgets[len(c.widgets)-1]
		}
		c.widgets[i-1].SetFocused(true)
		return c.widgets[i-1]
	}
	return nil
}

func (c *SimpleFocusChain) FocusDefault() Widget {
	if len(c.widgets) == 0 {
		return nil
	}
	return c.widgets[0]
}
