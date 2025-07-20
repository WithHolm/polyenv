package cmd

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"log/slog"
// 	"os"
// 	"path/filepath"
// 	"strings"

// 	"github.com/withholm/polyenv/internal/tools"
// 	"github.com/withholm/polyenv/internal/vaults"

// 	"github.com/spf13/cobra"
// )

// type PullOutputType string

// const (
// 	file         PullOutputType = "file"
// 	terminal     PullOutputType = "term"   // mykey=value
// 	terminaljson PullOutputType = "json"   // {"mykey":"value"}
// 	termJsonKv   PullOutputType = "jsonkv" // {"key":"mykey","value":"value"}
// )

// var pullPath string
// var pullOutput string
// var pullOutputType = terminal
// var allPullOutputTypes = []PullOutputType{file, terminal, terminaljson}

// var PullCmd = &cobra.Command{
// 	Use:   "pull",
// 	Short: "pull all secrets from keyvault",
// 	Long: `
// 		pull all secrets from keyvault.
// 		Any pull will also override existing .env file.
// 	`,
// 	Run: pull,
// }

// func init() {
// 	// PullCmd.Flags().VarP(&pullOutputType, "out", "o", "where to post the results of the pull. file|term|termjson")
// 	rootCmd.AddCommand(PullCmd)
// }

// // execute envault pull
// func pull(cmd *cobra.Command, args []string) {
// 	slog.Debug("pull called", "args", args)

// 	if len(args) > 0 {
// 		pullPath = args[0]
// 	} else {
// 		pullPath = Path
// 	}

// 	if pullPath != "" {
// 		err := tools.CheckDoubleDashS(pullPath, "path")
// 		if err != nil {
// 			log.Fatal(err.Error())
// 			os.Exit(1)
// 		}

// 		// in case they set '--path env.vaultopts'
// 		if pullPath == tools.GetVaultFilePath(pullPath) {
// 			log.Fatal("--path cannot be set to the vault options file")
// 			os.Exit(1)
// 		}
// 	}

// 	// get absolute path
// 	if !filepath.IsAbs(pullPath) {
// 		// path is absolute
// 		_path, err := filepath.Abs(pullPath)
// 		if err != nil {
// 			log.Fatal("failed to get absolute path: " + err.Error())
// 			os.Exit(1)
// 		}
// 		pullPath = _path
// 	}
// 	slog.Debug("absolute", "path", pullPath)

// 	// open vaultfile
// 	vaultFile, err := vaults.OpenVaultFile(pullPath)
// 	if err != nil {
// 		slog.Error("failed to open vault file: " + err.Error())
// 		os.Exit(1)
// 	}

// 	s, e := json.MarshalIndent(vaultFile, "", "  ")
// 	if e != nil {
// 		slog.Error("failed to marshal vault file: " + e.Error())
// 		os.Exit(1)
// 	}
// 	slog.Debug("vault file", "vaultFile", string(s))

// }

// // func outputSecrets(secrets map[string]string, outputType PullOutputType) error {
// // 	switch outputType {
// // 	case terminal:
// // 		for key, value := range secrets {
// // 			fmt.Println(key + "=" + value)
// // 		}
// // 	case terminaljson:
// // 		json, err := json.Marshal(secrets)
// // 		if err != nil {
// // 			return fmt.Errorf("failed to marshal json: %s", err)
// // 		}
// // 		fmt.Println(string(json))
// // 	case file:
// // 		//make file if it doesnt exist
// // 		if _, err := os.Stat(pullPath); os.IsNotExist(err) {
// // 			slog.Debug(fmt.Sprintf("creating .env file at %s", pullPath))
// // 			err := os.WriteFile(pullPath, []byte{}, 0644)
// // 			if err != nil {
// // 				return fmt.Errorf("failed to create .env file: %s", err)
// // 			}
// // 		}

// // 		err := godotenv.Write(secrets, pullPath)
// // 		if err != nil {
// // 			return fmt.Errorf("failed to write .env file: %s", err)
// // 		}
// // 	}
// // 	return nil
// // }

// // outputs type as string
// func (out *PullOutputType) String() string {
// 	return string(*out)
// }

// // sets output type
// func (out *PullOutputType) Set(value string) error {
// 	var errmgs = make([]string, 0)
// 	for _, typ := range allPullOutputTypes {
// 		if typ == PullOutputType(value) {
// 			*out = PullOutputType(value)
// 			return nil
// 		}
// 		// append error message, if none is found
// 		errmgs = append(errmgs, string(typ))
// 	}

// 	//return error if not found
// 	return fmt.Errorf("invalid output type: %s, must be %s", value, strings.Join(errmgs, ", "))
// }

// // returns output type
// func (out *PullOutputType) Type() string {
// 	return "outputType"
// }
