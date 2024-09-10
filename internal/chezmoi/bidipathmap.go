package chezmoi

import "github.com/twpayne/chezmoi/v2/internal/chezmoimaps"

type BiDiPathMap struct {
	forward *PathMap
	reverse *PathMap
}

func NewBiDiPathMap() *BiDiPathMap {
	return &BiDiPathMap{
		forward: NewPathMap(),
		reverse: NewPathMap(),
	}
}

func (m *BiDiPathMap) Add(fromRelPath, toRelPath RelPath) error {
	if err := m.forward.Add(fromRelPath, toRelPath); err != nil {
		return err
	}
	if err := m.reverse.Add(toRelPath, fromRelPath); err != nil {
		return err
	}
	return nil
}

func (m *BiDiPathMap) AddStringMap(sourceToTarget map[string]string) error {
	for _, fromRelPath := range chezmoimaps.SortedKeys(sourceToTarget) {
		if err := m.Add(NewRelPath(fromRelPath), NewRelPath(sourceToTarget[fromRelPath])); err != nil {
			return err
		}
	}
	return nil
}

func (m *BiDiPathMap) Forward(fromRelPath RelPath) RelPath {
	return m.forward.Lookup(fromRelPath)
}

func (m *BiDiPathMap) Reverse(toRelPath RelPath) RelPath {
	return m.reverse.Lookup(toRelPath)
}
