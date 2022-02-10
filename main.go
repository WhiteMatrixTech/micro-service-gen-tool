package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/google/uuid"
	"github.com/lwnmengjing/micro-service-gen-tool/pkg"
	"github.com/mitchellh/go-homedir"
)

var defaultTemplate = "git@github.com:Reimia/matrix-microservice-template.git"

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
	repo := prompt.Input(fmt.Sprintf("(default: %s: ", defaultTemplate),
		emptyCompleter)
	if repo == "" {
		repo = defaultTemplate
	}
	fmt.Println("Your input: ", repo)
	home, err := homedir.Dir()
	privateKeyFile := filepath.Join(home, ".ssh", "id_rsa")
	if err != nil {
		log.Fatalln(err)
	}
	templateWorkspace := filepath.Join("/tmp", uuid.New().String())
	fmt.Printf("git clone start: %s \n", time.Now().String())
	err = pkg.GitCloneSSH(repo, templateWorkspace, "main", privateKeyFile, "")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("git clone end: %s \n", time.Now().String())
	os.RemoveAll(filepath.Join(templateWorkspace, ".git"))
	//defer os.RemoveAll(templateWorkspace)
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
		projectName = "default"
	}
	keys, err := pkg.GetParseFromTemplate(filepath.Join(templateWorkspace, subPath))
	if err != nil {
		log.Fatalln(err)
	}

	for key := range keys {
		keys[key] = prompt.Input(fmt.Sprintf("template params %s: ", key), emptyCompleter)
	}

	err = pkg.Generate(&pkg.TemplateConfig{
		TemplateLocal: filepath.Join(templateWorkspace, subPath),
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
