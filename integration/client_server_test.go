// +build integration

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aquasecurity/trivy/internal"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestClientServer(t *testing.T) {
	type args struct {
		Version       string
		IgnoreUnfixed bool
		Severity      []string
		IgnoreIDs     []string
		Input         string
		ClientToken   string
		ServerToken   string
	}
	cases := []struct {
		name     string
		testArgs args
		golden   string
		wantErr  string
	}{
		{
			name: "alpine 3.10 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/alpine-310.tar.gz",
			},
			golden: "testdata/alpine-310.json.golden",
		},
		{
			name: "alpine 3.10 integration with token",
			testArgs: args{
				Version:     "dev",
				Input:       "testdata/fixtures/alpine-310.tar.gz",
				ClientToken: "token",
				ServerToken: "token",
			},
			golden: "testdata/alpine-310.json.golden",
		},
		{
			name: "alpine 3.10 integration with --ignore-unfixed option",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Input:         "testdata/fixtures/alpine-310.tar.gz",
			},
			golden: "testdata/alpine-310-ignore-unfixed.json.golden",
		},
		{
			name: "alpine 3.10 integration with medium and high severity",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Severity:      []string{"MEDIUM", "HIGH"},
				Input:         "testdata/fixtures/alpine-310.tar.gz",
			},
			golden: "testdata/alpine-310-medium-high.json.golden",
		},
		{
			name: "alpine 3.10 integration with .trivyignore",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: false,
				IgnoreIDs:     []string{"CVE-2019-1549", "CVE-2019-1563"},
				Input:         "testdata/fixtures/alpine-310.tar.gz",
			},
			golden: "testdata/alpine-310-ignore-cveids.json.golden",
		},
		{
			name: "alpine 3.9 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/alpine-39.tar.gz",
			},
			golden: "testdata/alpine-39.json.golden",
		},
		{
			name: "debian buster integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/debian-buster.tar.gz",
			},
			golden: "testdata/debian-buster.json.golden",
		},
		{
			name: "debian buster integration with --ignore-unfixed option",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Input:         "testdata/fixtures/debian-buster.tar.gz",
			},
			golden: "testdata/debian-buster-ignore-unfixed.json.golden",
		},
		{
			name: "debian stretch integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/debian-stretch.tar.gz",
			},
			golden: "testdata/debian-stretch.json.golden",
		},
		{
			name: "ubuntu 18.04 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/ubuntu-1804.tar.gz",
			},
			golden: "testdata/ubuntu-1804.json.golden",
		},
		{
			name: "ubuntu 18.04 integration with --ignore-unfixed option",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Input:         "testdata/fixtures/ubuntu-1804.tar.gz",
			},
			golden: "testdata/ubuntu-1804-ignore-unfixed.json.golden",
		},
		{
			name: "ubuntu 16.04 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/ubuntu-1604.tar.gz",
			},
			golden: "testdata/ubuntu-1604.json.golden",
		},
		{
			name: "centos 7 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/centos-7.tar.gz",
			},
			golden: "testdata/centos-7.json.golden",
		},
		{
			name: "centos 7 integration with --ignore-unfixed option",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Input:         "testdata/fixtures/centos-7.tar.gz",
			},
			golden: "testdata/centos-7-ignore-unfixed.json.golden",
		},
		{
			name: "centos 7 integration with critical severity",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Severity:      []string{"CRITICAL"},
				Input:         "testdata/fixtures/centos-7.tar.gz",
			},
			golden: "testdata/centos-7-critical.json.golden",
		},
		{
			name: "centos 7 integration with low and high severity",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Severity:      []string{"LOW", "HIGH"},
				Input:         "testdata/fixtures/centos-7.tar.gz",
			},
			golden: "testdata/centos-7-low-high.json.golden",
		},
		{
			name: "centos 6 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/centos-6.tar.gz",
			},
			golden: "testdata/centos-6.json.golden",
		},
		{
			name: "ubi 7 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/ubi-7.tar.gz",
			},
			golden: "testdata/ubi-7.json.golden",
		},
		{
			name: "distroless base integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/distroless-base.tar.gz",
			},
			golden: "testdata/distroless-base.json.golden",
		},
		{
			name: "distroless base integration with --ignore-unfixed option",
			testArgs: args{
				Version:       "dev",
				IgnoreUnfixed: true,
				Input:         "testdata/fixtures/distroless-base.tar.gz",
			},
			golden: "testdata/distroless-base-ignore-unfixed.json.golden",
		},
		{
			name: "distroless python27 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/distroless-python27.tar.gz",
			},
			golden: "testdata/distroless-python27.json.golden",
		},
		{
			name: "amazon 1 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/amazon-1.tar.gz",
			},
			golden: "testdata/amazon-1.json.golden",
		},
		{
			name: "amazon 2 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/amazon-2.tar.gz",
			},
			golden: "testdata/amazon-2.json.golden",
		},
		{
			name: "oracle 6 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/oraclelinux-6-slim.tar.gz",
			},
			golden: "testdata/oraclelinux-6-slim.json.golden",
		},
		{
			name: "oracle 7 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/oraclelinux-7-slim.tar.gz",
			},
			golden: "testdata/oraclelinux-7-slim.json.golden",
		},
		{
			name: "oracle 8 integration",
			testArgs: args{
				Version: "dev",
				Input:   "testdata/fixtures/oraclelinux-8-slim.tar.gz",
			},
			golden: "testdata/oraclelinux-8-slim.json.golden",
		},
		{
			name: "invalid token",
			testArgs: args{
				Version:     "dev",
				Input:       "testdata/fixtures/distroless-base.tar.gz",
				ClientToken: "invalidtoken",
				ServerToken: "token",
			},
			wantErr: "twirp error unauthenticated: invalid token",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Copy DB file
			cacheDir := gunzipDB()
			defer os.RemoveAll(cacheDir)

			port, err := getFreePort()
			require.NoError(t, err, c.name)
			addr := fmt.Sprintf("localhost:%d", port)

			go func() {
				// Setup CLI App
				app := internal.NewApp(c.testArgs.Version)
				app.Writer = ioutil.Discard
				osArgs := setupServer(addr, c.testArgs.ServerToken, cacheDir)

				// Run Trivy server
				require.NoError(t, app.Run(osArgs), c.name)
			}()

			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			require.NoError(t, waitPort(ctx, addr), c.name)

			// Setup CLI App
			app := internal.NewApp(c.testArgs.Version)
			app.Writer = ioutil.Discard

			osArgs, outputFile, cleanup := setupClient(t, c.testArgs.IgnoreUnfixed, c.testArgs.Severity,
				c.testArgs.IgnoreIDs, addr, c.testArgs.ClientToken, c.testArgs.Input, cacheDir, c.golden)
			defer cleanup()

			// Run Trivy client
			err = app.Run(osArgs)

			if c.wantErr != "" {
				require.NotNil(t, err, c.name)
				assert.Contains(t, err.Error(), c.wantErr, c.name)
				return
			} else {
				assert.NoError(t, err, c.name)
			}

			// Compare want and got
			want, err := ioutil.ReadFile(c.golden)
			assert.NoError(t, err)
			got, err := ioutil.ReadFile(outputFile)
			assert.NoError(t, err)

			assert.JSONEq(t, string(want), string(got))
		})
	}
}

func setupServer(addr, token, cacheDir string) []string {
	osArgs := []string{"trivy", "server", "--skip-update", "--cache-dir", cacheDir, "--listen", addr}
	if token != "" {
		osArgs = append(osArgs, []string{"--token", token}...)
	}
	return osArgs
}

func setupClient(t *testing.T, ignoreUnfixed bool, severity, ignoreIDs []string,
	addr, token, input, cacheDir, golden string) ([]string, string, func()) {
	t.Helper()
	osArgs := []string{"trivy", "client", "--cache-dir", cacheDir,
		"--format", "json", "--remote", "http://" + addr}
	if ignoreUnfixed {
		osArgs = append(osArgs, "--ignore-unfixed")
	}
	if len(severity) != 0 {
		osArgs = append(osArgs,
			[]string{"--severity", strings.Join(severity, ",")}...,
		)
	}

	var err error
	var ignoreTmpDir string
	if len(ignoreIDs) != 0 {
		ignoreTmpDir, err = ioutil.TempDir("", "ignore")
		require.NoError(t, err, "failed to create a temp dir")
		trivyIgnore := filepath.Join(ignoreTmpDir, ".trivyignore")
		err = ioutil.WriteFile(trivyIgnore, []byte(strings.Join(ignoreIDs, "\n")), 0444)
		require.NoError(t, err, "failed to write .trivyignore")
		osArgs = append(osArgs, []string{"--ignorefile", trivyIgnore}...)
	}
	if token != "" {
		osArgs = append(osArgs, []string{"--token", token}...)
	}
	if input != "" {
		osArgs = append(osArgs, []string{"--input", input}...)
	}

	// Setup the output file
	var outputFile string
	if *update {
		outputFile = golden
	} else {
		output, _ := ioutil.TempFile("", "integration")
		assert.Nil(t, output.Close())
		outputFile = output.Name()
	}

	cleanup := func() {
		_ = os.Remove(ignoreTmpDir)
		_ = os.Remove(outputFile)
	}

	osArgs = append(osArgs, []string{"--output", outputFile}...)
	return osArgs, outputFile, cleanup
}
