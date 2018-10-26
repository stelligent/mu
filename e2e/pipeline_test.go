// +build !unit

package e2e

import (
	"bufio"
	"bytes"
	hm "crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/stelligent/mu/common"
	mu_aws "github.com/stelligent/mu/provider/aws"
	"github.com/stelligent/mu/workflows"
	"github.com/termie/go-shutil"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/yaml.v2"
)

var contexts []*common.Context

// TestMain accepts a type *testing.M to override the primary main function to allow
// custom setup and tear down code. Eventually, m.Run() is called to invoke all
// tests of the general syntax (func Test*).
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

	includeTests := strings.Split(os.Getenv("INCLUDE_TESTS"), ",")
	if len(includeTests) == 1 && includeTests[0] == "" {
		includeTests = make([]string, 0)
	}

	var retCode int
	err = setupContexts(sess, dir, includeTests)
	if err != nil {
		fmt.Println(err)
		retCode = 1
	} else {
		// this invokes all func Test* functions in this file
		retCode = m.Run()
	}

	if retCode == 0 {
		fmt.Println("E2E test succeeded, cleaning up resources")
		teardownContexts(sess)
		os.RemoveAll(dir)
	} else {
		fmt.Println("E2E tests encountered an error, skipping cleaning up to allow debugging")
	}

	os.Exit(retCode)
}

func TestPipelineE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}
	var wg sync.WaitGroup
	for _, ctx := range contexts {
		fmt.Printf("scenario: %s\n", ctx.Config.Repo.Name)

		wg.Add(1)
		go func(c *common.Context) {
			err := validatePipeline(c)
			if err != nil {
				t.Errorf("Error on pipeline '%s': %v", c.Config.Repo.Name, err)
			}
			wg.Done()
		}(ctx)
	}

	wg.Wait()
}

func validatePipeline(ctx *common.Context) error {
	// - pipeline up
	err := workflows.NewPipelineUpserter(ctx, nil)()
	if err != nil {
		return err
	}

	// - wait for pipe success...
	err = waitForPipeline(ctx)
	if err != nil {
		return err
	}

	// - pipeline term
	err = workflows.NewPipelineTerminator(ctx, "")()
	if err != nil {
		fmt.Printf("Error on cleanup pipeline '%s': %v", ctx.Config.Repo.Name, err)
	}

	// - env term
	envNames := make([]string, 0)
	for _, env := range ctx.Config.Environments {
		envNames = append(envNames, env.Name)
	}

	err = workflows.NewEnvironmentsTerminator(ctx, envNames)()
	if err != nil {
		fmt.Printf("Error on cleanup envs: %v", err)
	}

	return nil
}

func waitForPipeline(ctx *common.Context) error {
	isRunning := true
	for isRunning {
		time.Sleep(30 * time.Second)

		pipelineName := fmt.Sprintf("%s-%s", ctx.Config.Namespace, ctx.Config.Repo.Name)
		states, err := ctx.PipelineManager.ListState(pipelineName)
		if err != nil {
			return err
		}

		status := codepipeline.ActionExecutionStatusInProgress
		latestExecutionID := ""
		latestActionTime := int64(0)
		for _, state := range states {
			if state.LatestExecution == nil {
				status = codepipeline.ActionExecutionStatusInProgress
				break
			} else if latestExecutionID == "" {
				latestExecutionID = common.StringValue(state.LatestExecution.PipelineExecutionId)
			} else if latestExecutionID != common.StringValue(state.LatestExecution.PipelineExecutionId) {
				status = codepipeline.ActionExecutionStatusInProgress
				break
			}

			for _, action := range state.ActionStates {
				if action.LatestExecution == nil {
					status = codepipeline.ActionExecutionStatusInProgress
					break
				} else {

					if latestActionTime <= aws.TimeValue(action.LatestExecution.LastStatusChange).Unix() {
						latestActionTime = aws.TimeValue(action.LatestExecution.LastStatusChange).Unix()
						status = common.StringValue(action.LatestExecution.Status)
					} else {
						status = codepipeline.ActionExecutionStatusInProgress
						break
					}

					fmt.Printf("  stage:%s action:%s status:%s\n",
						common.StringValue(state.StageName),
						common.StringValue(action.ActionName),
						status)

					if status == codepipeline.ActionExecutionStatusFailed {
						message := ""
						if action.LatestExecution.ErrorDetails != nil {
							message = common.StringValue(action.LatestExecution.ErrorDetails.Message)
						}

						return fmt.Errorf("failed on stage:%s action:%s - %s",
							common.StringValue(state.StageName),
							common.StringValue(action.ActionName), message)
					} else if status == codepipeline.ActionExecutionStatusInProgress {
						break
					}
				}
			}
		}

		isRunning = status == codepipeline.ActionExecutionStatusInProgress
	}

	return nil
}

func setupContexts(sess *session.Session, basedir string, includeTests []string) error {
	contexts = make([]*common.Context, 0)
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			if len(includeTests) > 0 {
				found := false
				for _, includeTest := range includeTests {
					if includeTest == f.Name() {
						found = true
						break
					}
				}

				if !found {
					fmt.Printf("Skipping test '%s' since not in list '%v'\n", f.Name(), includeTests)
					continue
				}
			}

			// create temp dir for repo
			dst := fmt.Sprintf("%s/%s", basedir, f.Name())
			err = shutil.CopyTree(f.Name(), dst, nil)
			if err != nil {
				fmt.Printf("Failure copying repo '%s'\n", f.Name())
				return err
			}

			// set the mu version and download URL
			muPath := fmt.Sprintf("%s/mu.yml", dst)
			err = updateMuYaml(muPath, os.Getenv("MU_BASEURL"), os.Getenv("MU_VERSION"))

			if err != nil {
				return err
			}

			// create a codecommit repo
			err = setupRepo(sess, f.Name(), dst)
			if err != nil {
				return err
			}

			// load mu.yml
			ctx, err := initContext(muPath)
			if err != nil {
				return err
			}

			contexts = append(contexts, ctx)
		}
	}

	return nil
}

func updateMuYaml(muConfigFile string, muBaseurl string, muVersion string) error {
	config := new(common.Config)
	yamlFile, err := os.Open(muConfigFile)

	fmt.Printf("Updating %s with baseurl=%s and version=%s\n", muConfigFile, muBaseurl, muVersion)

	yamlBuffer := new(bytes.Buffer)
	yamlBuffer.ReadFrom(bufio.NewReader(yamlFile))
	err = yaml.Unmarshal(yamlBuffer.Bytes(), config)
	if err != nil {
		return err
	}

	config.Service.Pipeline.MuBaseurl = muBaseurl
	config.Service.Pipeline.MuVersion = muVersion

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(muConfigFile, configBytes, 0600)
	if err != nil {
		return err
	}

	return nil
}

func initContext(muConfigFile string) (*common.Context, error) {
	// initialize context
	context := common.NewContext()
	namespace := os.Getenv("MU_NAMESPACE")
	if namespace != "" {
		context.Config.Namespace = namespace
	} else {
		context.Config.Namespace = "mu"
	}
	err := context.InitializeContext()
	if err != nil {
		return nil, err
	}

	err = mu_aws.InitializeContext(context, "", "", "", "", true, "", false)
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
	fmt.Printf("Setting up repo %s from %s\n", name, dir)

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
		URLs: []string{
			common.StringValue(resp.RepositoryMetadata.CloneUrlHttp),
		},
	})
	if err != nil {
		fmt.Printf("Failure create remote '%s'\n", name)
		return err
	}

	username, password := credentialHelper(sess, common.StringValue(resp.RepositoryMetadata.CloneUrlHttp))

	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}
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
