package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/c-bata/go-prompt"
	"github.com/google/uuid"
	"github.com/lwnmengjing/micro-service-gen-tool/pkg"
	"github.com/mitchellh/go-homedir"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Parsed()
}

func emptyCompleter(_ prompt.Document) []prompt.Suggest {
	s := make([]prompt.Suggest, 0)
	return s
}

func subPathCompleter(sub []string) prompt.Completer {
	s := make([]prompt.Suggest, len(sub))
	for i := range sub {
		s[i] = prompt.Suggest{
			Text:        sub[i],
			Description: fmt.Sprintf("select template %s", sub[i]),
		}
	}
	return func(in prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
	}
}

func main() {
	repo := prompt.Input("(default: git@github.com:lwnmengjing/template-demo.git: ",
		emptyCompleter)
	if repo == "" {
		repo = "git@github.com:lwnmengjing/template-demo.git"
	}
	fmt.Println("Your input: ", repo)
	home, err := homedir.Dir()
	privateKeyFile := filepath.Join(home, ".ssh", "id_rsa")
	if err != nil {
		log.Fatalln(err)
	}
	templateWorkspace := filepath.Join("/tmp", uuid.New().String())
	err = pkg.GitCloneSSH(repo, templateWorkspace, "main", privateKeyFile, "")
	if err != nil {
		log.Fatalln(err)
	}
	os.RemoveAll(filepath.Join(templateWorkspace, ".git"))
	defer os.RemoveAll(templateWorkspace)
	sub, err := pkg.GetSubPath(templateWorkspace)
	if err != nil {
		log.Fatalln(err)
	}
	if len(sub) == 0 {
		log.Fatalln("not found template")
	}
	subPath := prompt.Input(fmt.Sprintf("select template(default:%s): ", sub[0]), subPathCompleter(sub))
	if subPath == "" {
		subPath = sub[0]
	}
	projectName := prompt.Input("project name(default: default): ", emptyCompleter)
	if projectName == "" {
		log.Fatalln("project name can't empty")
	}
	keys, err := pkg.GetParseFromTemplate(filepath.Join(templateWorkspace, subPath))
	if err != nil {
		log.Fatalln(err)
	}

	for key := range keys {
		keys[key] = prompt.Input(fmt.Sprintf("template params %s >>>", key), emptyCompleter)
	}
	fmt.Println(keys)

	err = pkg.Generate(&pkg.TemplateConfig{
		TemplateLocal: templateWorkspace,
		CreateRepo:    false,
		Destination:   filepath.Join(".", projectName),
		Github:        nil,
		Params:        keys,
		Ignore:        nil,
	})
	if err != nil {
		log.Fatalln(err)
	}
}
