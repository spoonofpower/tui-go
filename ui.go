package tui

import "os"

type UI interface {
	SetTheme(p *Theme)
	SetKeybinding(k interface{}, fn func())
	SetFocusChain(ch FocusChain)
	Run(*os.File) error
	Quit()
}

func New(root Widget, term string) (UI, error) {
	tcellui, err := newTcellUI(root, term)
	if err != nil {
		return nil, err
	}
	return tcellui, nil
}
