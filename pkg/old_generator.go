/*
 * @Author: lwnmengjing
 * @Date: 2021/12/16 7:39 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/12/16 7:39 下午
 */

package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/zealic/xignore"
)

type TemplateConfig struct {
	Service       string        `yaml:"service"`
	TemplateUrl   string        `yaml:"templateUrl"`
	TemplateLocal string        `yaml:"templateLocal"`
	CreateRepo    bool          `yaml:"createRepo"`
	Destination   string        `yaml:"destination"`
	Github        *GithubConfig `yaml:"github"`
	Params        interface{}   `yaml:"params"`
	Ignore        []string      `yaml:"ignore"`
}

func (e *TemplateConfig) OnChange() {
	fmt.Println("config changed")
}

// OldGenerator generate operator
type OldGenerator struct {
	TemplatePath             string
	DestinationPath          string
	Cfg                      interface{}
	GithubConfig             *GithubConfig
	accessToken              string
	TemplateIgnoreDirs       []string
	TemplateIgnoreFiles      []string
	TemplateParseIgnoreDirs  []string
	TemplateParseIgnoreFiles []string
}

// OldGenerate example
//func OldGenerate(url, destinationPath string, cfg interface{}, githubConfig *GithubConfig, accessToken string) error {
//	templatePath := filepath.Base(url)
func OldGenerate(c *TemplateConfig) (err error) {
	var templatePath string
	var accessToken string
	if c.Github != nil {
		accessToken = c.Github.Token
	}
	if c.TemplateLocal != "" {
		templatePath = c.TemplateLocal
	} else {
		templatePath = filepath.Base(c.TemplateUrl)
		_ = os.RemoveAll(templatePath)
		err = GitClone(c.TemplateUrl, templatePath, false, accessToken)
		if err != nil {
			log.Println(err)
			return err
		}
		defer os.RemoveAll(templatePath)
	}

	if !c.CreateRepo {
		c.Github = nil
	}
	//delete destinationPath
	_ = os.RemoveAll(c.Destination)
	_ = os.RemoveAll(filepath.Join(templatePath, ".git"))
	templateResultIgnore, err := xignore.DirMatches(templatePath, &xignore.MatchesOptions{
		Ignorefile: TemplateIgnore,
		Nested:     true, // Handle nested ignorefile
	})
	if err != nil {
		log.Println(err)
		return err
	}
	//_ = os.RemoveAll(filepath.Join(templatePath, TemplateIgnore))
	templateParseResultIgnore, err := xignore.DirMatches(templatePath, &xignore.MatchesOptions{
		Ignorefile: TemplateParseIgnore,
		Nested:     true,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	//_ = os.RemoveAll(filepath.Join(templatePath, TemplateParseIgnore))
	t := &OldGenerator{
		TemplatePath:             templatePath,
		DestinationPath:          c.Destination,
		Cfg:                      c.Params,
		GithubConfig:             c.Github,
		accessToken:              accessToken,
		TemplateIgnoreDirs:       templateResultIgnore.MatchedDirs,
		TemplateIgnoreFiles:      templateResultIgnore.MatchedFiles,
		TemplateParseIgnoreDirs:  templateParseResultIgnore.MatchedDirs,
		TemplateParseIgnoreFiles: templateParseResultIgnore.MatchedFiles,
	}
	if err = t.Traverse(); err != nil {
		log.Println(err)
		return err
	}
	if err = t.CreateGithubRepo(); err != nil {
		log.Println(err)
		return err
	}
	if err = t.CommitGithubRepo(); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Traverse traverse all dir
func (e *OldGenerator) Traverse() error {
	return filepath.WalkDir(e.TemplatePath, e.TraverseFunc)
}

// TraverseFunc traverse callback
func (e *OldGenerator) TraverseFunc(path string, f os.DirEntry, err error) error {
	switch filepath.Base(path) {
	case TemplateIgnore, TemplateParseIgnore:
		return nil
	}
	// template ignore
	if len(e.TemplateIgnoreDirs) > 0 {
		for i := range e.TemplateIgnoreDirs {
			if f.IsDir() && strings.Index(path, filepath.Join(e.TemplatePath, e.TemplateIgnoreDirs[i])) == 0 {
				return nil
			}
		}
	}
	if len(e.TemplateIgnoreFiles) > 0 {
		for i := range e.TemplateIgnoreFiles {
			if filepath.Join(e.TemplatePath, e.TemplateIgnoreFiles[i]) == path {
				return nil
			}
		}
	}
	templatePath := path
	t := template.New(path)
	t = template.Must(t.Parse(path))
	var buffer bytes.Buffer
	if err = t.Execute(&buffer, e.Cfg); err != nil {
		log.Println(err)
		return err
	}
	path = strings.ReplaceAll(buffer.String(), e.TemplatePath, e.DestinationPath)

	if f.IsDir() {
		// dir
		if !PathExist(path) {
			return PathCreate(path)
		}
		return nil
	}
	var parseIgnore bool
	// template parse ignore
	if len(e.TemplateParseIgnoreDirs) > 0 {
		for i := range e.TemplateParseIgnoreDirs {
			if strings.Index(templatePath, filepath.Join(e.TemplatePath, e.TemplateParseIgnoreDirs[i])) == 0 {
				parseIgnore = true
			}
		}
	}
	if !parseIgnore && len(e.TemplateParseIgnoreFiles) > 0 {
		for i := range e.TemplateParseIgnoreFiles {
			if filepath.Join(e.TemplatePath, e.TemplateParseIgnoreFiles[i]) == templatePath {
				parseIgnore = true
			}
		}
	}
	if parseIgnore {
		_, err = FileCopy(templatePath, path)
		if err != nil {
			log.Println(err)
		}
		return err
	}
	var rb []byte
	if rb, err = ioutil.ReadFile(templatePath); err != nil {
		log.Println(err)
		return err
	}
	buffer = bytes.Buffer{}
	t = template.New(path + "[file]")
	t = template.Must(t.Parse(string(rb)))
	if err = t.Execute(&buffer, e.Cfg); err != nil {
		log.Println(err)
		return err
	}
	fi, err := f.Info()
	if err != nil {
		log.Println(err)
		return err
	}
	// create file
	err = FileCreate(buffer, path, fi.Mode())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// CreateGithubRepo create github repo
func (e *OldGenerator) CreateGithubRepo() error {
	if e.GithubConfig == nil {
		return nil
	}
	repo, err := CreateGithubRepo(
		e.GithubConfig.Organization,
		e.GithubConfig.Name,
		e.GithubConfig.Description,
		e.accessToken, true)
	if err != nil {
		log.Println(err)
		return err
	}
	//owner := e.GithubConfig.Organization
	//if repo.Organization == nil || *repo.Organization.Name == "" {
	//	owner = *repo.Owner.Login
	//}
	//err = AddActionSecretsGithubRepo(
	//	owner,
	//	e.GithubConfig.Name,
	//	e.accessToken, e.GithubConfig.Secrets)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}
	err = PathCreate(e.DestinationPath)
	if err != nil {
		log.Println(err)
		return err
	}
	err = GitRemote(repo.GetCloneURL(), e.DestinationPath)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// CommitGithubRepo commit repo to github
func (e *OldGenerator) CommitGithubRepo() (err error) {
	if e.GithubConfig == nil {
		return nil
	}
	return CommitAndPushGithubRepo(e.DestinationPath, e.accessToken)
}
