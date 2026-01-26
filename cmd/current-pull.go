// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
)

func generatePullCommand() *cobra.Command {
	var pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "pull all defined secrets from vaults",
		Long: `
		pull all secrets from vault.
		if option.use dot secret file is enabled it will be set in .env.secret file
		if not, it will try to find a key set in any of your .env files. if it cannot find it, it will error out.
	`,
		RunE: pull,
	}
	return pullCmd
}

//region !pullfunc

// pull all defined secrets from vaults
func pull(cmd *cobra.Command, args []string) error {
	secretFilename := PolyenvFile.GenerateFileName(".env.secret")
	var secretFilePath string
	existingEnv, err := PolyenvFile.AllDotenvValues()
	if err != nil {
		return fmt.Errorf("failed to get existing env: %w", err)
	}

	//region pull:precheck
	//precheck. if not using .secret file, check if key exists in existing dotenv file
	// TODO: Is this neccessary? how can we know where to create the env key-val pair?
	slog.Debug("precheck!", "use dot secret file", PolyenvFile.Options.UseDotSecretFileForSecrets)
	if !PolyenvFile.Options.UseDotSecretFileForSecrets {
		for k := range PolyenvFile.Secrets {
			matches := 0
			for _, f := range existingEnv {
				if f.Key == k {
					matches++
				}
			}
			if matches == 0 {
				return fmt.Errorf("opted out of .secret creation and cannot find a existing reference: %s", k)
			} else if matches > 1 {
				return fmt.Errorf("there are multiple references to the same key in .env files. please remove all but one: %s", k)
			}
		}
	} else {
		root, e := tools.GetGitRootOrCwd()
		if e != nil {
			return fmt.Errorf("failed to get project root: %w", e)
		}

		secretFiles, e := tools.GetAllFiles(root, []string{secretFilename}, tools.MatchNameIExact)
		if e != nil {
			return fmt.Errorf("failed to get files: %w", e)
		}
		if len(secretFiles) > 1 {
			return fmt.Errorf("multiple .env.secret files found; expected exactly one: %s", secretFiles)
		} else if len(secretFiles) == 0 {
			secretFilePath = filepath.Join(root, secretFilename)
			// create new file
			if err := os.WriteFile(secretFilePath, []byte{}, 0o600); err != nil {
				return fmt.Errorf("failed to create .env.secret file: %w", err)
			}
		} else {
			secretFilePath = secretFiles[0]
		}
	}

	//region pull:from vaults
	cnt := 0
	contents := make([]model.StoredEnv, 0)
	for k, v := range PolyenvFile.Secrets {
		cnt++
		spn := spinner.Points
		spn.FPS = time.Second / 15
		sp := spinner.New().Title(k + " pulling").Type(spn)
		err := sp.ActionWithErr(func(ctx context.Context) error {
			prefix := fmt.Sprintf(" %d/%d - %s -> ", cnt, len(PolyenvFile.Secrets), k)
			// get vault from file
			slog.Debug("pulling", "secret", v.RemoteKey, "vault", v.Vault)
			sp = sp.Title(prefix + " getting local vault definition")
			vlt, ok := PolyenvFile.Vaults[v.Vault]
			if !ok {
				return fmt.Errorf("vault not found: %s", v.Vault)
			}

			//warming up
			sp = sp.Title(prefix + " warming up")
			err := vlt.Warmup()
			if err != nil {
				return fmt.Errorf("failed to warmup vault: %w", err)
			}

			// elevate permissions
			sp = sp.Title(prefix + " elevating permissions to " + v.Vault)
			err = vlt.PullElevate()
			if err != nil {
				return fmt.Errorf("failed to elevate permissions: %w", err)
			}

			//pulling secret
			sp = sp.Title(prefix + " pulling from " + v.Vault)
			content, err := vlt.Pull(v)
			if err != nil {
				return fmt.Errorf("failed to pull secret: %w", err)
			}
			contents = append(contents, model.StoredEnv{
				Value: string(content.Value.Bytes()),
				Key:   v.LocalKey,
			})
			return nil
		}).Run()
		if err != nil {
			return fmt.Errorf("failed to run pull action: %w", err)
		}
	}

	//region pull:write files
	for _, newEnv := range contents {
		if PolyenvFile.Options.UseDotSecretFileForSecrets {
			slog.Debug("writing to .env.secret", "key", newEnv.Key, "file", secretFilePath)
			// slog.Info("writing to .env.secret", "key", newEnv.Key, "file", secretFilePath)
			newEnv.File = secretFilePath
			e := newEnv.Save()
			if e != nil {
				return fmt.Errorf("failed to write to .env.secret: %w", e)
			}
			continue
		}

		for _, v := range existingEnv {
			if v.Key == newEnv.Key {
				slog.Debug("updating existing env", "key", v.Key, "file", v.File)
				v.Value = newEnv.Value
				e := v.Save()
				if e != nil {
					return fmt.Errorf("failed to update existing env: %w", e)
				}
				break
			}
		}
	}
	return nil
}
