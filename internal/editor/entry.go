package editor

/*
====================
  Entry Points
====================
*/

func OpenFile(path string) error {
	ed, err := newEditorFromFile(path)
	if err != nil {
		return err
	}
	return ed.run()
}

func OpenProject(path string) error {
	ed, err := newEditorFromProject(path)
	if err != nil {
		return err
	}
	return ed.run()
}
