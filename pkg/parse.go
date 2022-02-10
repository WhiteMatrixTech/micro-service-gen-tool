package pkg

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template/parse"

	"github.com/zealic/xignore"
)

var TemplateIgnore = ".templateignore"
var TemplateParseIgnore = ".templateparseignore"

// getParseKeys get parse keys from template text
func getParseKeys(nodes *parse.ListNode) []string {
	keys := make([]string, 0)
	if nodes == nil {
		return keys
	}
	for a := range nodes.Nodes {
		if actionNode, ok := nodes.Nodes[a].(*parse.ActionNode); ok {
			if actionNode == nil || actionNode.Pipe == nil {
				continue
			}
			for b := range actionNode.Pipe.Cmds {
				if strings.Index(actionNode.Pipe.Cmds[b].String(), ".") == 0 {
					keys = append(keys, actionNode.Pipe.Cmds[b].String()[1:])
				}
			}
		}
	}
	return keys
}

// GetParseFromTemplate get parse keys from template
func GetParseFromTemplate(dir string) (map[string]string, error) {
	keys := make(map[string]string, 0)
	templateResultIgnore, err := xignore.DirMatches(filepath.Join(dir, TemplateIgnore),
		&xignore.MatchesOptions{
			Ignorefile: TemplateIgnore,
			Nested:     true, // Handle nested ignorefile
		})
	if err != nil && err != os.ErrNotExist {
		log.Println(err)
		return nil, err
	}
	templateParseResultIgnore, err := xignore.DirMatches(filepath.Join(dir, TemplateParseIgnore),
		&xignore.MatchesOptions{
			Ignorefile: TemplateParseIgnore,
			Nested:     true,
		})
	if err != nil && err != os.ErrNotExist {
		log.Println(err)
		return nil, err
	}
	ignoreDirs := make([]string, 0)
	ignoreFiles := make([]string, 0)
	if templateResultIgnore != nil {
		ignoreDirs = templateResultIgnore.MatchedDirs
		ignoreFiles = templateResultIgnore.MatchedFiles
	}
	if templateParseResultIgnore != nil {
		ignoreDirs = templateParseResultIgnore.MatchedDirs
		ignoreFiles = templateParseResultIgnore.MatchedFiles
	}
	err = filepath.WalkDir(dir, parseTraverse(dir, keys, ignoreDirs, ignoreFiles))
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func parseTraverse(dir string, keys map[string]string, ignoreDirs, ignoreFiles []string) fs.WalkDirFunc {
	if keys == nil {
		keys = make(map[string]string)
	}
	return func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			for i := range ignoreDirs {
				if strings.Index(path, filepath.Join(dir, ignoreDirs[i])) == 0 {
					return nil
				}
			}
		} else {
			for i := range ignoreFiles {
				if filepath.Join(dir, ignoreFiles[i]) == path {
					return nil
				}
			}
		}
		{
			tree, err := parse.Parse("path", path, "{{", "}}")
			if err != nil {
				return err
			}
			if tree == nil {
				return nil
			}
			for _, key := range getParseKeys(tree["path"].Root) {
				keys[key] = ""
			}
		}
		if d.IsDir() {
			return nil
		}
		rb, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		{
			tree, err := parse.Parse("file", string(rb), "{{", "}}")
			if err != nil {
				return err
			}
			if tree == nil {
				return nil
			}
			for _, key := range getParseKeys(tree["file"].Root) {
				keys[key] = ""
			}
		}
		return nil
	}
}
