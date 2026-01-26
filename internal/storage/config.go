package storage

import (
	"fmt"
	"log/slog"

	"github.com/BurntSushi/toml"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
)

type ConfigRepository struct {
	File *model.File
}

func NewFileRepository(root string) (*ConfigRepository, error) {
	//read/open file

	allfiles, err := tools.GetAllFiles(root, []string{"polyenv.toml"}, tools.MatchNameIExact)
	if err != nil {
		return nil, err
	}
	if len(allfiles) == 0 {
		return nil, fmt.Errorf("no polyenvfile found")
	}
	if len(allfiles) > 1 {
		for _, f := range allfiles {
			slog.Error("found polyenvfile", "path", f)
		}
		return nil, fmt.Errorf("multiple polyenvfiles found in '%s'. can only have one.", root)
	}

	path := allfiles[0]
	var file *model.File
	meta, err := toml.DecodeFile(path, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read polyenvfile: %w", err)
	}

	if len(meta.Undecoded()) > 0 {
		slog.Warn("got undecoded items in polyenvfile", "undecoded", meta.Undecoded())
	}

	return &ConfigRepository{
		File: file,
	}, nil
}

func (r *ConfigRepository) Save() error {
	return nil
}

func (r *ConfigRepository) GetVault(name string) (model.Vault, error) {
	return nil, nil
}

func (r *ConfigRepository) GetVaultNames() []string {
	return nil
}
