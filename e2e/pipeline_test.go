// +build !unit

package e2e

import (
	"bytes"
	hm "crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/provider/aws"
	"github.com/stretchr/testify/assert"
	"github.com/termie/go-shutil"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

var contexts []*common.Context

func TestMain(m *testing.M) {
	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}
	sess, err := session.NewSessionWithOptions(sessOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dir, err := ioutil.TempDir("", "mu-e2e-repos")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var retCode int
	err = setupContexts(sess, dir)
	if err != nil {
		fmt.Println(err)
		retCode = 1
	} else {
		retCode = m.Run()
	}

	teardownContexts(sess)
	os.RemoveAll(dir)
	os.Exit(retCode)
}

func TestPipelineE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}
	assert := assert.New(t)

	for _, ctx := range contexts {
		fmt.Printf("scenario: %s\n", ctx.Config.Repo.Name)

		// - pipeline up
		// - wait...
		// - pipeline term
		// - env term
	}
	assert.Fail("force fail")
}

func setupContexts(sess *session.Session, basedir string) error {
	contexts = make([]*common.Context, 0)
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			// create temp dir for repo
			dst := fmt.Sprintf("%s/%s", basedir, f.Name())
			err = shutil.CopyTree(f.Name(), dst, nil)
			if err != nil {
				fmt.Printf("Failure copying repo '%s'\n", f.Name())
				return err
			}

			// create a codecommit repo
			err = setupRepo(sess, f.Name(), dst)
			if err != nil {
				return err
			}

			// load mu.yml
			ctx, err := initContext(fmt.Sprintf("%s/mu.yml", dst))
			if err != nil {
				return err
			}

			contexts = append(contexts, ctx)
		}
	}

	return nil
}

func initContext(muConfigFile string) (*common.Context, error) {
	// initialize context
	context := common.NewContext()
	err := context.InitializeContext()
	if err != nil {
		return nil, err
	}

	err = aws.InitializeContext(context, "", "", false)
	if err != nil {
		return nil, err
	}

	err = context.InitializeConfigFromFile(muConfigFile)
	if err != nil {
		return nil, err
	}

	return context, nil
}

func teardownContexts(sess *session.Session) {
	for _, ctx := range contexts {
		err := teardownRepo(sess, ctx)
		if err != nil {
			fmt.Printf("Unable to teardown repo: %s", err)
		}
	}
}
func setupRepo(sess *session.Session, name string, dir string) error {
	// create codecommit repo
	ccAPI := codecommit.New(sess)
	ccAPI.DeleteRepository(&codecommit.DeleteRepositoryInput{
		RepositoryName: &name,
	})
	resp, err := ccAPI.CreateRepository(&codecommit.CreateRepositoryInput{
		RepositoryName: &name,
	})
	if err != nil {
		fmt.Printf("Failure creating codecommit repo '%s'\n", name)
		return err
	}

	// git init
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		fmt.Printf("Failure init repo '%s'\n", name)
		return err
	}

	// git commit
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = gitAddAll(dir, ".", wt)
	if err != nil {
		fmt.Printf("Failure add '%s'\n", name)
		return err
	}
	_, err = wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "mu-e2e",
			Email: "mu@stelligent.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		fmt.Printf("Failure commit '%s'\n", name)
		return err
	}

	// git push
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URL:  common.StringValue(resp.RepositoryMetadata.CloneUrlHttp),
	})
	if err != nil {
		fmt.Printf("Failure create remote '%s'\n", name)
		return err
	}

	username, password := credentialHelper(sess, common.StringValue(resp.RepositoryMetadata.CloneUrlHttp))

	auth := http.NewBasicAuth(username, password)
	err = remote.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	if err != nil {
		fmt.Printf("Failure pushing '%s'\n", name)
		return err
	}

	return nil

}

func gitAddAll(gitdir string, subdir string, wt *git.Worktree) error {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", gitdir, subdir))
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.Name() == ".git" {
			continue
		}
		prefixedName := filepath.Clean(fmt.Sprintf("%s/%s", subdir, f.Name()))
		if f.IsDir() {
			err = gitAddAll(gitdir, prefixedName, wt)
			if err != nil {
				return err
			}
		} else {
			_, err = wt.Add(prefixedName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func teardownRepo(sess *session.Session, ctx *common.Context) error {
	// delete codecommit repo
	ccAPI := codecommit.New(sess)
	params := &codecommit.DeleteRepositoryInput{
		RepositoryName: &ctx.Config.Repo.Name,
	}
	_, err := ccAPI.DeleteRepository(params)
	if err != nil {
		return err
	}

	return nil
}

var regexpCC = regexp.MustCompile(`git-codecommit\.([^.]+)\.amazonaws\.com.*`)

func credentialHelper(sess *session.Session, giturl string) (string, string) {
	u, _ := url.Parse(giturl)

	t := time.Now().UTC()
	//t, _ := time.Parse("20060102T150405", "20170824T072504") // reference time the comments were generated with

	creds, _ := sess.Config.Credentials.Get()
	username := creds.AccessKeyID
	if creds.SessionToken != "" {
		username = fmt.Sprintf("%s%s%s", username, "%", creds.SessionToken)
	}

	region := common.StringValue(sess.Config.Region)

	if match := regexpCC.FindStringSubmatch(u.Host); len(match) > 1 {
		region = match[1]
	}

	// Build canonical request
	// 'GIT\n/v1/repos/httpd\n\nhost:git-codecommit.us-east-1.amazonaws.com\n\nhost\n'
	cr := new(bytes.Buffer)
	fmt.Fprintf(cr, "%s\n", "GIT")         // HTTPRequestMethod
	fmt.Fprintf(cr, "%s\n", u.Path)        // CanonicalURI
	fmt.Fprintf(cr, "%s\n", "")            // CanonicalQueryString
	fmt.Fprintf(cr, "host:%s\n\n", u.Host) // CanonicalHeaders
	fmt.Fprintf(cr, "%s\n", "host")        // SignedHeaders
	fmt.Fprintf(cr, "%s", "")              // HexEncode(Hash(Payload))

	// Build string to sign
	// 'AWS4-HMAC-SHA256\n20160227T172057\n20160227/us-east-1/codecommit/aws4_request\n650b9e2de2abce7c30f6ad51c4a84b361e1f8aaaa3152e93d35509450db2d869'
	sts := new(bytes.Buffer)
	fmt.Fprint(sts, "AWS4-HMAC-SHA256\n")                                                   // Algorithm
	fmt.Fprintf(sts, "%s\n", t.Format("20060102T150405"))                                   // RequestDate
	fmt.Fprintf(sts, "%s/%s/%s/aws4_request\n", t.Format("20060102"), region, "codecommit") // CredentialScope
	fmt.Fprintf(sts, "%s", hash(cr.String()))                                               // HexEncode(Hash(CanonicalRequest))

	// Calculate the AWS Signature Version 4
	// '56238a5ac75c8bd36bba91737377aca46c867f584a6695f0486f3e3bba9b4ed5'
	dsk := hmac([]byte("AWS4"+creds.SecretAccessKey), []byte(t.Format("20060102")))
	dsk = hmac(dsk, []byte(region))
	dsk = hmac(dsk, []byte("codecommit"))
	dsk = hmac(dsk, []byte("aws4_request"))
	h := hmac(dsk, []byte(sts.String()))
	sig := fmt.Sprintf("%x", h) // HexEncode(HMAC(derived-signing-key, string-to-sign))

	// codecommmit smart http password to use with AWS_ACCESS
	password := fmt.Sprintf("%sZ%s", t.Format("20060102T150405"), sig)
	//fmt.Printf("url:%s username:%s password:%s\n", u, username, password)

	return username, password
}

// hash method calculates the sha256 hash for a given string
func hash(in string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s", in)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// hmac method calculates the sha256 hmac for a given slice of bytes
func hmac(key, data []byte) []byte {
	h := hm.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
