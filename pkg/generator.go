/*
 * @Author: lwnmengjing
 * @Date: 2021/12/16 7:39 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/12/16 7:39 下午
 */

package pkg

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/zealic/xignore"
)

// Generator generate operator
type Generator struct {
	SubPath                  string
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

// Generate example
//func Generate(url, destinationPath string, cfg interface{}, githubConfig *GithubConfig, accessToken string) error {
//	templatePath := filepath.Base(url)
func Generate(c *TemplateConfig) (err error) {
	var templatePath string
	var accessToken string
	//if c.Github != nil {
	//	accessToken = c.Github.Token
	//}
	if c.TemplateLocal != "" {
		templatePath = c.TemplateLocal
	} else {
		templatePath = filepath.Base(c.TemplateUrl)
		//_ = os.RemoveAll(templatePath)
		//err = GitClone(c.TemplateUrl, templatePath, false, accessToken)
		//if err != nil {
		//	log.Println(err)
		//	return err
		//}
		//defer os.RemoveAll(templatePath)
	}
	subPath := filepath.Join(templatePath, c.Service)

	if !c.CreateRepo {
		c.Github = nil
	}
	//delete destinationPath
	_ = os.RemoveAll(c.Destination)

	t := &Generator{
		SubPath:                  c.Service,
		TemplatePath:             templatePath,
		DestinationPath:          c.Destination,
		Cfg:                      c.Params,
		GithubConfig:             c.Github,
		accessToken:              accessToken,
		TemplateIgnoreDirs:       make([]string, 0),
		TemplateIgnoreFiles:      make([]string, 0),
		TemplateParseIgnoreDirs:  make([]string, 0),
		TemplateParseIgnoreFiles: make([]string, 0),
	}

	{
		templateResultIgnore, err := xignore.DirMatches(templatePath, &xignore.MatchesOptions{
			Ignorefile: TemplateIgnore,
			Nested:     true, // Handle nested ignorefile
		})
		if err != nil && err != os.ErrNotExist {
			log.Println(err)
			return err
		}
		if templateResultIgnore != nil {
			t.TemplateIgnoreDirs = templateResultIgnore.MatchedDirs
			t.TemplateIgnoreFiles = templateResultIgnore.MatchedFiles
		}
		templateParseResultIgnore, err := xignore.DirMatches(templatePath, &xignore.MatchesOptions{
			Ignorefile: TemplateParseIgnore,
			Nested:     true,
		})
		if err != nil && err != os.ErrNotExist {
			log.Println(err)
			return err
		}
		if templateParseResultIgnore != nil {
			t.TemplateParseIgnoreDirs = templateParseResultIgnore.MatchedDirs
			t.TemplateParseIgnoreFiles = templateParseResultIgnore.MatchedFiles
		}
		_ = os.RemoveAll(filepath.Join(templatePath, TemplateParseIgnore))
	}

	{
		templateResultIgnore, err := xignore.DirMatches(subPath, &xignore.MatchesOptions{
			Ignorefile: TemplateIgnore,
			Nested:     true, // Handle nested ignorefile
		})
		if err != nil && err != os.ErrNotExist {
			log.Println(err)
			return err
		}
		if templateResultIgnore != nil {

			for i := range templateResultIgnore.MatchedDirs {
				t.TemplateIgnoreDirs = append(t.TemplateIgnoreDirs,
					strings.Join(strings.Split(templateResultIgnore.MatchedDirs[i], "/")[1:], "/"))
			}
			for i := range templateResultIgnore.MatchedDirs {
				t.TemplateIgnoreFiles = append(t.TemplateIgnoreFiles,
					strings.Join(strings.Split(templateResultIgnore.MatchedDirs[i], "/")[1:], "/"))
			}
			//t.TemplateIgnoreDirs = templateResultIgnore.MatchedDirs
			//t.TemplateIgnoreFiles = templateResultIgnore.MatchedFiles
		}
		//_ = os.RemoveAll(filepath.Join(templatePath, TemplateIgnore))
		templateParseResultIgnore, err := xignore.DirMatches(subPath, &xignore.MatchesOptions{
			Ignorefile: TemplateParseIgnore,
			Nested:     true,
		})
		if err != nil && err != os.ErrNotExist {
			log.Println(err)
			return err
		}
		if templateParseResultIgnore != nil {
			t.TemplateParseIgnoreDirs = append(t.TemplateParseIgnoreDirs, templateParseResultIgnore.MatchedDirs...)
			t.TemplateParseIgnoreFiles = append(t.TemplateParseIgnoreFiles, templateParseResultIgnore.MatchedFiles...)
		}
		_ = os.RemoveAll(filepath.Join(subPath, TemplateParseIgnore))
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
func (e *Generator) Traverse() error {
	return filepath.WalkDir(filepath.Join(e.TemplatePath, e.SubPath), e.TraverseFunc)
}

// TraverseFunc traverse callback
func (e *Generator) TraverseFunc(path string, f os.DirEntry, err error) error {
	switch filepath.Base(path) {
	case TemplateIgnore, TemplateParseIgnore:
		return nil
	}
	// template ignore
	if len(e.TemplateIgnoreDirs) > 0 {
		for i := range e.TemplateIgnoreDirs {
			if f.IsDir() &&
				(strings.Index(path, filepath.Join(e.TemplatePath, e.TemplateIgnoreDirs[i])) == 0 ||
					strings.Index(path, filepath.Join(e.TemplatePath, e.SubPath, e.TemplateIgnoreDirs[i])) == 0) {
				return nil
			}
		}
	}
	if len(e.TemplateIgnoreFiles) > 0 {
		for i := range e.TemplateIgnoreFiles {
			if filepath.Join(e.TemplatePath, e.TemplateIgnoreFiles[i]) == path ||
				filepath.Join(e.TemplatePath, e.SubPath, e.TemplateIgnoreFiles[i]) == path {
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
	path = strings.ReplaceAll(buffer.String(), filepath.Join(e.TemplatePath, e.SubPath), e.DestinationPath)

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
			if strings.Index(templatePath, filepath.Join(e.TemplatePath, e.TemplateParseIgnoreDirs[i])) == 0 ||
				strings.Index(templatePath, filepath.Join(e.SubPath, e.TemplatePath, e.TemplateParseIgnoreDirs[i])) == 0 {
				parseIgnore = true
			}
		}
	}
	if !parseIgnore && len(e.TemplateParseIgnoreFiles) > 0 {
		for i := range e.TemplateParseIgnoreFiles {
			if filepath.Join(e.TemplatePath, e.TemplateParseIgnoreFiles[i]) == templatePath ||
				filepath.Join(e.SubPath, e.TemplatePath, e.TemplateParseIgnoreFiles[i]) == templatePath {
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
func (e *Generator) CreateGithubRepo() error {
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
func (e *Generator) CommitGithubRepo() (err error) {
	if e.GithubConfig == nil {
		return nil
	}
	return CommitAndPushGithubRepo(e.DestinationPath, e.accessToken)
}
