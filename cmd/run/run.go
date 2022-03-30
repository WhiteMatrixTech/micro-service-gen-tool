package run

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/mss-boot-io/micro-service-gen-tool/pkg"
)

var (
	StartCmd = &cobra.Command{
		Use:     "run",
		Short:   "Start generate project",
		Example: "generate-tool run",
		PreRun: func(cmd *cobra.Command, args []string) {
			pkg.Upgrade(true)
			pre()

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	// If using repoGitCloneSSHWithPrompts, use this value
	// defaultTemplate = "git@github.com:WhiteMatrixTech/matrix-microservice-template.git"

	// If using repoGitCloneViaDeployerAccount, use this value
	defaultTemplate = "https://github.com/WhiteMatrixTech/matrix-microservice-template.git"
)

func pre() {
	if os.Getenv("MICRO_DEFAULT_TEMPLATE") != "" {
		defaultTemplate = os.Getenv("MICRO_DEFAULT_TEMPLATE")
	}
}

/*
	SSH Git, not use now in favor of git S3 git token way
	TODO may consider add back if there is a better way of using it
*/
func repoGitCloneSSHWithPrompts(repo, templateWorkspace, branch string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	var password string

	// private ID
	privateID := "id_ed25519"
	fmt.Printf("private key id (default: %s): ", pkg.Yellow(privateID))
	_, _ = fmt.Scanf("%s", &privateID)
	privateKeyFile := filepath.Join(home, ".ssh", privateID)

	// password
	fmt.Printf("private pem password (default: %s): ", pkg.Yellow("''"))
	_, _ = fmt.Scanf("%s", &password)
	fmt.Printf("git clone start: %s \n", time.Now().String())
	fmt.Println(privateKeyFile)

	// do the git clone
	return pkg.GitCloneSSH(repo, templateWorkspace, branch, privateKeyFile, password)
}

func repoGitCloneViaDeployerAccount(repo, templateWorkspace, branch string) error {
	token, err := pkg.ReadTokenFromS3()
	if err != nil {
		fmt.Println(pkg.Red("Failed to read the Github token from S3, please make sure you have correct aws credentials set up."))
		return err
	}
	return pkg.GitCloneViaDeployerAccount(repo, templateWorkspace, branch, token)
}

func run() error {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	var err error

	// template repo
	repo := defaultTemplate
	fmt.Printf("template repo (default: %s): ", pkg.Yellow(defaultTemplate))
	_, _ = fmt.Scanf("%s", &repo)
	fmt.Println("your template repo: ", pkg.Cyan(repo))

	// git branch
	branch := "main"
	fmt.Printf("template repo branch (default: '%s'): ", pkg.Yellow(branch))
	_, _ = fmt.Scanf("%s", &branch)

	// workspace file location
	templateWorkspace := "/tmp/template-workspace"
	fmt.Printf("template workspace (default: %s): ", pkg.Yellow(templateWorkspace))
	_, _ = fmt.Scanf("%s", &templateWorkspace)
	templateWorkspace = filepath.Join(templateWorkspace, uuid.New().String())

	// TODO we can replace repoGitCloneViaDeployerAccount with the SSH one below
	if strings.Index(repo, "@") > 0 {
		err = repoGitCloneSSHWithPrompts(repo, templateWorkspace, branch)
	} else {
		err = repoGitCloneViaDeployerAccount(repo, templateWorkspace, branch)
	}

	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("git clone end: %s \n", time.Now().String())
	_ = os.RemoveAll(filepath.Join(templateWorkspace, ".git"))
	defer os.RemoveAll(templateWorkspace)
	sub, err := pkg.GetSubPath(templateWorkspace)
	if err != nil {
		log.Fatalln(err)
	}
	if len(sub) == 0 {
		log.Fatalln("Template not found!")
	}
	fmt.Println(pkg.Yellow("************* please select sub path ***************"))
	for i := range sub {
		fmt.Println(pkg.Yellow("* "), pkg.Yellow(sub[i]))
	}
	fmt.Println(pkg.Yellow("****************************************************"))
SUBPATH:
	subPath := sub[0]
	fmt.Printf("select template sub path (default: %s): ", pkg.Cyan(subPath))
	_, _ = fmt.Scanf("%s", &subPath)
	ok := false
	for i := range sub {
		if sub[i] == subPath {
			ok = true
		}
	}
	if !ok {
		fmt.Println(pkg.Red("please select exist sub path."))
		goto SUBPATH
	}
	projectName := "default"
	fmt.Printf("project name(default:%s)", pkg.Yellow(projectName))
	_, _ = fmt.Scanf("%s", &projectName)
	keys, err := pkg.GetParseFromTemplate(templateWorkspace, subPath)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(pkg.Magenta("please config your param's value"))

	for key := range keys {
		var value string
	BACK:
		fmt.Printf("%s: ", pkg.Yellow(key))
		_, _ = fmt.Scanf("%s", &value)
		if value == "" {
			goto BACK
		}
		keys[key] = value
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
	fmt.Println(pkg.Green("template generate project success...."))
	return nil
}
