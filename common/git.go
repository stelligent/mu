package common

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-ini/ini"
	"gopkg.in/yaml.v2"
)

func findGitRevision(file string) (string, error) {
	gitDir, err := findGitDirectory(file)
	if err != nil {
		return "", err
	}

	head, err := findGitHead(file)
	if err != nil {
		return "", err
	}
	// load commitid ref
	refBuf, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", gitDir, head))
	if err != nil {
		return "", err
	}
	return string(string(refBuf)[:7]), nil
}

func findGitBranch(file string) (string, error) {
	head, err := findGitHead(file)
	if err != nil {
		return "", err
	}

	// get branch name
	branch := strings.TrimPrefix(head, "refs/heads/")
	log.Debugf("Found branch: %s", branch)
	return branch, nil
}

func findGitHead(file string) (string, error) {
	gitDir, err := findGitDirectory(file)
	if err != nil {
		return "", err
	}
	log.Debugf("Loading revision from git directory '%s'", gitDir)

	// load HEAD ref
	headFile, err := os.Open(fmt.Sprintf("%s/HEAD", gitDir))
	if err != nil {
		return "", err
	}
	defer func() {
		headFile.Close()
	}()

	headBuffer := new(bytes.Buffer)
	headBuffer.ReadFrom(bufio.NewReader(headFile))
	head := make(map[string]string)
	yaml.Unmarshal(headBuffer.Bytes(), head)

	log.Debugf("HEAD points to '%s'", head["ref"])

	return head["ref"], nil
}

func findGitRemoteURL(file string) (string, error) {
	gitDir, err := findGitDirectory(file)
	if err != nil {
		return "", err
	}
	log.Debugf("Loading slug from git directory '%s'", gitDir)

	gitconfig, err := ini.InsensitiveLoad(fmt.Sprintf("%s/config", gitDir))
	if err != nil {
		return "", err
	}
	remote, err := gitconfig.GetSection("remote \"origin\"")
	if err != nil {
		return "", err
	}
	urlKey, err := remote.GetKey("url")
	if err != nil {
		return "", err
	}
	url := urlKey.String()
	return url, nil
}

func findGitSlug(url string) (string, string, error) {
	codeCommitHTTPRegex := regexp.MustCompile("^http(s?)://git-codecommit\\.(.+)\\.amazonaws.com/v1/repos/(.+)$")
	codeCommitSSHRegex := regexp.MustCompile("ssh://git-codecommit\\.(.+)\\.amazonaws.com/v1/repos/(.+)$")
	httpRegex := regexp.MustCompile("^http(s?)://.*github.com.*/(.+)/(.+).git$")
	sshRegex := regexp.MustCompile("github.com:(.+)/(.+).git$")

	if matches := codeCommitHTTPRegex.FindStringSubmatch(url); matches != nil {
		return "CodeCommit", matches[3], nil
	} else if matches := codeCommitSSHRegex.FindStringSubmatch(url); matches != nil {
		return "CodeCommit", matches[2], nil
	} else if matches := httpRegex.FindStringSubmatch(url); matches != nil {
		return "GitHub", fmt.Sprintf("%s/%s", matches[2], matches[3]), nil
	} else if matches := sshRegex.FindStringSubmatch(url); matches != nil {
		return "GitHub", fmt.Sprintf("%s/%s", matches[1], matches[2]), nil
	}
	return "", url, nil
}

func findGitDirectory(fromFile string) (string, error) {
	absPath, err := filepath.Abs(fromFile)
	if err != nil {
		return "", err
	}

	log.Debugf("Searching for git directory in %s", absPath)
	fi, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	var dir string
	if fi.Mode().IsDir() {
		dir = absPath
	} else {
		dir = path.Dir(absPath)
	}

	gitPath := path.Join(dir, ".git")
	fi, err = os.Stat(gitPath)
	if err == nil && fi.Mode().IsDir() {
		return gitPath, nil
	} else if dir == "/" || dir == "C:\\" || dir == "c:\\" {
		return "", errors.New("Unable to find git repo")
	}

	return findGitDirectory(filepath.Dir(dir))

}
