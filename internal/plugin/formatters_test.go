package plugin

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/withholm/polyenv/internal/model"
)

// update is a command-line flag to update the golden files.
// To use it, run: go test -v ./internal/plugin -update
var update = flag.Bool("update", false, "update golden files")

type formatterTestCase struct {
	name  string
	input []model.StoredEnv
}

func TestFormatters(t *testing.T) {
	testCases := []formatterTestCase{
		{
			name: "basic",
			input: []model.StoredEnv{
				{Key: "A", Value: "1", IsSecret: false},
				{Key: "B", Value: "2", IsSecret: true},
			},
		},
		// {
		// 	name: "different values",
		// 	input: []model.StoredEnv{
		// 		{Key: "A", Value: "1", IsSecret: false},
		// 		{Key: "B", Value: 1, IsSecret: true},
		// 		{Key: "C", Value: "3", IsSecret: false},
		// 	},
		// }
		// Add more shared test cases here in the future.
	}

	type formatterTest struct {
		name      string
		formatter model.Formatter
	}

	formatters := []formatterTest{
		{name: "dotenv", formatter: &DotenvFormatter{}},
		{name: "json", formatter: &JSONFormatter{AsArray: false}},
		{name: "json_array", formatter: &JSONFormatter{AsArray: true}},
		{name: "azdevops", formatter: &AzDevopsFormatter{}},
		{name: "pwsh", formatter: &PwshFormatter{}},
		{name: "passthrough", formatter: &PassthroughFormatter{}},
		// Note: 'pick' and 'stats' formatters are excluded as they are special cases.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, f := range formatters {
				t.Run(f.name, func(t *testing.T) {
					goldenFile := filepath.Join("testdata", tc.name+"."+f.name+".golden")

					gotBytes, err := f.formatter.OutputFormat(tc.input)
					if err != nil {
						t.Fatalf("OutputFormat() returned an error: %v", err)
					}

					// Windows generates CRLF, so we normalize to LF for consistent comparisons.
					got := strings.ReplaceAll(string(gotBytes), "\r\n", "\n")
					// t.Log("got:", got)
					if *update {
						t.Logf("updating golden file: %s", goldenFile)
						if err := os.WriteFile(goldenFile, []byte(got), 0644); err != nil {
							t.Fatalf("failed to update golden file: %v", err)
						}
					}

					wantBytes, err := os.ReadFile(goldenFile)
					if err != nil {
						t.Fatalf("failed to read golden file: %v", err)
					}
					want := strings.ReplaceAll(string(wantBytes), "\r\n", "\n")

					if diff := cmp.Diff(want, got); diff != "" {
						t.Errorf("OutputFormat() mismatch (-want +got):\n%s", diff)
					}
				})
			}
		})
	}
}
