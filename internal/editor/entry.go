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
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config loading fails, use default config
		cfg = config.DefaultConfig()
	}

	ed, err := newEditorFromFile(path, cfg)
	if err != nil {
		return err
	}
	return ed.run()
}

func OpenProject(path string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config loading fails, use default config
		cfg = config.DefaultConfig()
	}

	ed, err := newEditorFromProject(path, cfg)
	if err != nil {
		return err
	}
	return ed.run()
}
