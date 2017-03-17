package common

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

func TestFindGitSlug(t *testing.T) {
	assert := assert.New(t)

	var slugTests = []struct {
		url      string // input
		provider string // expected result
		slug     string // expected result
	}{
		{"https://git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo-name", "CodeCommit", "my-repo-name"},
		{"git@github.com:stelligent/mu.git", "GitHub", "stelligent/mu"},
		{"https://github.com/stelligent/mu.git", "GitHub", "stelligent/mu"},
		{"http://github.com/stelligent/mu.git", "GitHub", "stelligent/mu"},
		{"http://myotherrepo.com/mu.git", "", "http://myotherrepo.com/mu.git"},
	}

	for _, tt := range slugTests {

		basedir, err := ioutil.TempDir("", "mu-test")
		defer os.RemoveAll(basedir)

		assert.Nil(err)

		err = gitCmd("init", basedir)
		assert.Nil(err)

		err = gitCmd("config", "-f", fmt.Sprintf("%s/.git/config", basedir), "--add", "remote.origin.url", tt.url)
		assert.Nil(err)

		provider, slug, err := findGitSlug(basedir)

		assert.Nil(err)
		assert.Equal(tt.provider, provider)
		assert.Equal(tt.slug, slug)
	}

}

func gitCmd(args ...string) error {
	var stdout bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = ioutil.Discard

	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return fmt.Errorf("Exit error %d", waitStatus.ExitStatus())
		}
		return exitError
	}
	return nil
}
