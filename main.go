package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lwnmengjing/micro-service-gen-tool/pkg"
	"github.com/mitchellh/go-homedir"
)

var defaultTemplate = "git@github.com:WhiteMatrixTech/matrix-microservice-template.git"

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Parsed()
}

func main() {
	var err error
	//var repo string
	repo := defaultTemplate
	fmt.Printf("template repo (default:%s):", defaultTemplate)
	_, _ = fmt.Scanf("%s", &repo)
	fmt.Println("your template repo:", repo)
	branch := "main"
	fmt.Printf("template repo branch(default:'%s'):", branch)
	_, _ = fmt.Scanf("%s", &branch)
	home, err := homedir.Dir()
	privateID := "id_rsa"
	fmt.Printf("private key id(default:%s):", privateID)
	_, _ = fmt.Scanf("%s", &privateID)
	privateKeyFile := filepath.Join(home, ".ssh", privateID)
	if err != nil {
		log.Fatalln(err)
	}
	templateWorkspace := "/tmp/template-workspace"
	fmt.Printf("template workspace(default:%s):", templateWorkspace)
	_, _ = fmt.Scanf("%s", &templateWorkspace)
	templateWorkspace = filepath.Join(templateWorkspace, uuid.New().String())
	var password string
	fmt.Print("private pem password(default:''):")
	_, _ = fmt.Scanf("%s", &password)
	fmt.Printf("git clone start: %s \n", time.Now().String())
	err = pkg.GitCloneSSH(repo, templateWorkspace, branch, privateKeyFile, password)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("git clone end: %s \n", time.Now().String())
	os.RemoveAll(filepath.Join(templateWorkspace, ".git"))
	defer os.RemoveAll(templateWorkspace)
	sub, err := pkg.GetSubPath(templateWorkspace)
	if err != nil {
		log.Fatalln(err)
	}
	if len(sub) == 0 {
		log.Fatalln("not found template")
	}
	fmt.Println("please select sub path:")
	for i := range sub {
		fmt.Println(sub[i])
	}
SUBPATH:
	subPath := sub[0]
	fmt.Printf("select template sub path(default:%s):", subPath)
	_, _ = fmt.Scanf("%s", &subPath)
	ok := false
	for i := range sub {
		if sub[i] == subPath {
			ok = true
		}
	}
	if !ok {
		fmt.Println("please select exist sub path.")
		goto SUBPATH
	}
	projectName := "default"
	fmt.Printf("project name(default:%s)", projectName)
	_, _ = fmt.Scanf("%s", &projectName)
	keys, err := pkg.GetParseFromTemplate(filepath.Join(templateWorkspace, subPath))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("please config your param's value")

	for key := range keys {
		var value string
	BACK:
		fmt.Printf("%s:", key)
		_, _ = fmt.Scanf("%s", &value)
		if value == "" {
			goto BACK
		}
		keys[key] = value
		//keys[key] = prompt.Input(fmt.Sprintf("template params %s: ", key), emptyCompleter)
	}

	err = pkg.Generate(&pkg.TemplateConfig{
		Service:       subPath,
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
	fmt.Println("template generate project success....")
}
