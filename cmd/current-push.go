package cmd

// import (
// 	"fmt"
// 	"log/slog"
// 	"os"
// 	"path/filepath"
// 	"time"

// 	"github.com/charmbracelet/huh/spinner"
// 	"github.com/spf13/cobra"
// 	"github.com/withholm/polyenv/internal/model"
// 	"github.com/withholm/polyenv/internal/tools"
// )

// func generatePullCommand() *cobra.Command {
// 	var pullCmd = &cobra.Command{
// 		Use:   "pull",
// 		Short: "pull all defined secrets from vaults",
// 		Long: `
// 		pull all secrets from vault.
// 		if option.use dot secret file is enabled it will be set in .env.secret file
// 		if not, it will try to find a key set in any of your .env files. if it cannot find it, it will error out.
// 	`,
// 		Run: pull,
// 	}
// 	return pullCmd
// }

// //region !pullfunc

// // pull all defined secrets from vaults
// func pull(cmd *cobra.Command, args []string) {
// 	secretFilename := PolyenvFile.GenerateFileName(".env.secret")
// 	var secretFilePath string
// 	existingEnv, err := PolyenvFile.AllDotenvValues()
// 	if err != nil {
// 		slog.Error("failed to get existing env", "error", err)
// 		os.Exit(1)
// 	}

// 	//region pull:precheck
// 	//precheck. if not using .secret file, check if key exists in existing dotenv file
// 	// TODO: Is this neccessary? how can we know where to create the env key-val pair?
// 	slog.Debug("precheck!", "use dot secret file", PolyenvFile.Options.UseDotSecretFileForSecrets)
// 	if !PolyenvFile.Options.UseDotSecretFileForSecrets {
// 		for k := range PolyenvFile.Secrets {
// 			matches := 0
// 			for _, f := range existingEnv {
// 				if f.Key == k {
// 					matches++
// 				}
// 			}
// 			if matches == 0 {
// 				slog.Error("opted out of .secret creation and cannot find a existing reference", "key", k)
// 				os.Exit(1)
// 			} else if matches > 1 {
// 				slog.Error("there are multiple references to the same key in .env files. please remove all but one", "key", k)
// 				os.Exit(1)
// 			}
// 		}
// 	} else {
// 		root, e := tools.GetGitRootOrCwd()
// 		if e != nil {
// 			slog.Error("failed to get project root", "error", e)
// 			os.Exit(1)
// 		}

// 		secretFiles, e := tools.GetAllFiles(root, []string{secretFilename}, tools.MatchNameIExact)
// 		if e != nil {
// 			slog.Error("failed to get files", "error", e)
// 			os.Exit(1)
// 		}
// 		if len(secretFiles) > 1 {
// 			slog.Error("multiple .env.secret files found; expected exactly one", "files", secretFiles)
// 			os.Exit(1)
// 		} else if len(secretFiles) == 0 {
// 			secretFilePath = filepath.Join(root, secretFilename)
// 			// create new file
// 			if err := os.WriteFile(secretFilePath, []byte{}, 0o600); err != nil {
// 				slog.Error("failed to create .env.secret file", "error", err)
// 				os.Exit(1)
// 			}
// 		} else {
// 			secretFilePath = secretFiles[0]
// 		}
// 	}

// 	//region pull:from vaults
// 	cnt := 0
// 	contents := make([]model.StoredEnv, 0)
// 	for k, v := range PolyenvFile.Secrets {
// 		cnt++
// 		spn := spinner.Points
// 		spn.FPS = time.Second / 15
// 		sp := spinner.New().Title(k + " pulling").Type(spn)
// 		sp.Action(func() {
// 			prefix := fmt.Sprintf(" %d/%d - %s -> ", cnt, len(PolyenvFile.Secrets), k)
// 			// get vault from file
// 			slog.Debug("pulling", "secret", v.RemoteKey, "vault", v.Vault)
// 			sp = sp.Title(prefix + " getting local vault definition")
// 			vlt, ok := PolyenvFile.Vaults[v.Vault]
// 			if !ok {
// 				slog.Error("vault not found", "vault", v.Vault)
// 				os.Exit(1)
// 			}

// 			//warming up
// 			sp = sp.Title(prefix + " warming up")
// 			err := vlt.Warmup()
// 			if err != nil {
// 				slog.Error("failed to warmup vault", "vault", v.Vault, "error", err)
// 				os.Exit(1)
// 			}

// 			// elevate permissions
// 			sp = sp.Title(prefix + " elevating permissions to " + v.Vault)
// 			err = vlt.PullElevate()
// 			if err != nil {
// 				slog.Error("failed to elevate permissions", "vault", v.Vault, "error", err)
// 				os.Exit(1)
// 			}

// 			//pulling secret
// 			sp = sp.Title(prefix + " pulling from " + v.Vault)
// 			content, err := vlt.Pull(v)
// 			if err != nil {
// 				slog.Error("failed to pull secret", "vault", v.Vault, "error", err)
// 				os.Exit(1)
// 			}
// 			contents = append(contents, model.StoredEnv{
// 				Value: content.Value,
// 				Key:   v.LocalKey,
// 			})
// 		}).Run()
// 	}

// 	//region pull:write files
// 	for _, newEnv := range contents {
// 		if PolyenvFile.Options.UseDotSecretFileForSecrets {
// 			slog.Debug("writing to .env.secret", "key", newEnv.Key, "file", secretFilePath)
// 			// slog.Info("writing to .env.secret", "key", newEnv.Key, "file", secretFilePath)
// 			newEnv.File = secretFilePath
// 			e := newEnv.Save()
// 			if e != nil {
// 				slog.Error("failed to write to .env.secret", "error", e)
// 				os.Exit(1)
// 			}
// 			continue
// 		}

// 		for _, v := range existingEnv {
// 			if v.Key == newEnv.Key {
// 				slog.Debug("updating existing env", "key", v.Key, "file", v.File)
// 				v.Value = newEnv.Value
// 				e := v.Save()
// 				if e != nil {
// 					slog.Error("failed to update existing env", "error", e)
// 					os.Exit(1)
// 				}
// 				break
// 			}
// 		}
// 	}

// }
