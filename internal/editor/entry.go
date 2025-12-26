package editor

import (
	"github.com/dragonbytelabs/voidabyss/internal/config"
)

/*
====================
  Entry Points
====================
*/

func OpenFile(path string) error {
	cfg, loader, err := config.LoadConfig()
	if err != nil {
		// If config loading fails, use default config
		cfg = config.DefaultConfig()
	}

	ed, err := newEditorFromFile(path, cfg, loader)
	if err != nil {
		if loader != nil {
			loader.Close()
		}
		return err
	}
	defer func() {
		if loader != nil {
			loader.Close()
		}
	}()

	// Fire events after initialization
	ed.FireEditorReady()
	ed.FireBufRead()

	return ed.run()
}

func OpenProject(path string) error {
	cfg, loader, err := config.LoadConfig()
	if err != nil {
		// If config loading fails, use default config
		cfg = config.DefaultConfig()
	}

	ed, err := newEditorFromProject(path, cfg, loader)
	if err != nil {
		if loader != nil {
			loader.Close()
		}
		return err
	}
	defer func() {
		if loader != nil {
			loader.Close()
		}
	}()

	// Fire events after initialization
	ed.FireEditorReady()
	ed.FireBufRead()

	return ed.run()
}
