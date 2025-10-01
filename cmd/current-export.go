// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/plugin"
	"github.com/withholm/polyenv/internal/tools"
)

var writerFlag string
var formatFlag string

// export --to {writer} --as {format}

func generateEnvCommand() *cobra.Command {
	var envCmd = &cobra.Command{
		Use:   "export",
		Short: "export environment variables to a given format and destination",
		Long: `
		export environment variables to a given format and destination. defaults to json output to stdout
	`,
		Run: ExportEnv,
	}

	envCmd.Flags().StringVar(&writerFlag, "to", "stdout", fmt.Sprintf("where to output to: %v", tools.MapKeySlice(plugin.Writers)))
	err := envCmd.RegisterFlagCompletionFunc("to", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.MapKeySlice(plugin.Writers), cobra.ShellCompDirectiveKeepOrder
	})
	if err != nil {
		slog.Error("failed to add completion on 'to' flag", "err", err)
	}

	formatters := tools.MapKeySlice(plugin.OutputFormatters)
	formatters = append(formatters, "auto")

	envCmd.Flags().StringVar(&formatFlag, "as", "auto", fmt.Sprintf("how to format the output: %v", formatters))
	err = envCmd.RegisterFlagCompletionFunc("as", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return formatters, cobra.ShellCompDirectiveKeepOrder
	})
	if err != nil {
		slog.Error("failed to add completion on 'as' flag", "err", err)
	}

	return envCmd
}

func ExportEnv(cmd *cobra.Command, args []string) {
	list, err := PolyenvFile.AllDotenvValues()
	if err != nil {
		slog.Error("failed to list env", "error", err)
		os.Exit(1)
	}

	slog.Debug("output", "as", formatFlag, "to", writerFlag)

	wFunc, ok := tools.InequalFindInMap(plugin.Writers, writerFlag)
	if !ok {
		slog.Error("failed to find output writer", "to", writerFlag)
		os.Exit(1)
	}
	writer := wFunc()
	plugin.SelectedWriter = writer

	//detect format that can be used with the given writer
	if strings.EqualFold(formatFlag, "auto") {
		formatFlag, err = plugin.AutoOutputFormat(writer)
		if err != nil {
			slog.Error("failed to get auto format", "error", err)
			os.Exit(1)
		}
	}

	//validate that given format is accepted by writer
	if !plugin.AcceptsFormat(writer, formatFlag) {
		slog.Error("writer does not support format", "writer", writerFlag, "format", formatFlag)
		a, d := writer.AcceptedFormats()
		slog.Info("accepted formats", "writer", writerFlag, "formats", a)
		slog.Info("denied formats", "writer", writerFlag, "formats", d)
		os.Exit(1)
	}

	slog.Debug("writer supports format", "writer", writerFlag, "format", formatFlag)

	fmtFunc, ok := tools.InequalFindInMap(plugin.OutputFormatters, formatFlag)
	if !ok {
		slog.Error("failed to find formatter", "as", formatFlag)
		os.Exit(1)
	}

	formatter := fmtFunc()

	if strings.EqualFold(formatFlag, "stats") {
		formatter = &plugin.StatsFormatter{PolyenvFile: PolyenvFile}
	}

	//format output
	formatted, err := formatter.OutputFormat(list)
	if err != nil {
		slog.Error("failed to format output", "error", err)
		os.Exit(1)
	}

	//write output
	err = writer.Write(formatted)
	if err != nil {
		slog.Error("failed to write output", "error", err)
		os.Exit(1)
	}
}
