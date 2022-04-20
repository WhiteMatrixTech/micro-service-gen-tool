/*
 * @Author: lwnmengjing
 * @Date: 2021/12/16 9:07 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/12/16 9:07 下午
 */

package pkg

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v43/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type GithubConfig struct {
	Name         string            `yaml:"name"`
	Organization string            `yaml:"organization"`
	Repository   string            `yaml:"repository"`
	Branch       string            `yaml:"branch"`
	Description  string            `yaml:"description"`
	Secrets      map[string]string `yaml:"secrets"`
}

type GithubConstantValues struct {
	DefaultAccount      string
	DefaultOrganization string
	DefaultBranch       string
}

var GithubConstants = &GithubConstantValues{
	DefaultAccount:      "whitematrix-deployer",
	DefaultOrganization: "WhiteMatrixTech",
	DefaultBranch:       "main",
}

var githubToken = ""
var githubClient *github.Client = nil

func GetDefaultGithubToken() string {
	if githubToken == "" {
		token, err := ReadTokenFromS3()
		if err != nil {
			fmt.Println(Red("Failed to read the Github token from S3, please make sure you have correct aws credentials set up."))
			log.Fatal(err.Error())
		}
		githubToken = token
	}
	return githubToken
}

func GetDefaultGithubInstance() *github.Client {
	if githubClient == nil {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: GetDefaultGithubToken()},
		)
		tc := oauth2.NewClient(ctx, ts)

		githubClient = github.NewClient(tc)
	}
	return githubClient
}

// GitRemote from remote git
func GitRemote(url, directory string) error {
	r, err := git.PlainInit(directory, false)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		log.Println(err)
		return err
	}
	err = r.CreateBranch(&config.Branch{
		Name: "main",
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func CheckGithubRepoExistence(organization, repo string) (bool, error) {
	_, resp, err := GetDefaultGithubInstance().Repositories.Get(context.Background(), organization, repo)
	if resp != nil && resp.Response.StatusCode == 200 {
		return true, nil
	}
	if resp != nil && resp.Response.StatusCode == 404 {
		return false, nil
	}
	return false, err
}

// GitClone clone git repo
func GitCloneViaDeployerAccount(url, directory, reference, accessToken string) error {
	auth := &http.BasicAuth{}
	if accessToken != "" {
		auth.Username = GithubConstants.DefaultAccount // this is hardcoded
		auth.Password = accessToken
	}
	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		Auth:          auth,
		URL:           url,
		Progress:      os.Stdout,
		Depth:         1,
		ReferenceName: plumbing.NewBranchReferenceName(reference),
	})
	return err
}

// GitCloneSSH clone git repo from ssh
func GitCloneSSH(url, directory, reference, privateKeyFile, password string) error {
	_, err := os.Stat(privateKeyFile)
	if err != nil {
		return errors.Errorf("read file %s failed %s\n", privateKeyFile, err.Error())
	}
	publicKey, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, password)
	if err != nil {
		return errors.Errorf("generate publickeys failed: %s\n", err.Error())
	}
	_, err = git.PlainClone(directory, false, &git.CloneOptions{
		Auth:          publicKey,
		URL:           url,
		Progress:      os.Stdout,
		Depth:         1,
		ReferenceName: plumbing.NewBranchReferenceName(reference),
	})
	if err != nil {
		return errors.Errorf("clone repo error: %s", err.Error())
	}
	return nil
}

// CreateGithubRepo create github repo
func CreateGithubRepo(organization, name, description string, private bool) (*github.Repository, error) {

	r := &github.Repository{Name: &name, Private: &private, Description: &description}
	repo, _, err := GetDefaultGithubInstance().Repositories.Create(context.Background(), organization, r)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Printf("Successfully created new repo: %s\n", repo.GetName())
	return repo, nil
}

// AddActionSecretsGithubRepo add action secret
//func AddActionSecretsGithubRepo(organization, name, token string, data map[string]string) error {
//	ctx := context.Background()
//	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
//	tc := oauth2.NewClient(ctx, ts)
//	client := github.NewClient(tc)
//	var err error
//	for k, v := range data {
//		input := github.EncryptedSecret{
//			Name: k,
//			EncryptedValue: v,
//		}
//		_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, organization, name, &input)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

// CommitAndPushGithubRepo commit and push github repo
func CommitAndPushGithubRepo(directory, branchName string) error {
	r, err := git.PlainOpen(directory)
	if err != nil {
		log.Println(err)
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.Commit(":tada: added service package", &git.CommitOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("git push origin main:refs/heads/%s -f", branchName))
	cmd.Dir = directory
	_, err = cmd.Output()
	if err != nil {
		log.Printf("failed to push: %v\n", err)
		return err
	}

	return err
}
