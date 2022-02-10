package main

import (
	"flag"
	"github.com/lwnmengjing/core-go/config"
	"github.com/lwnmengjing/core-go/config/source/file"
	"log"
	"path/filepath"

	"github.com/lwnmengjing/micro-service-gen-tool/pkg"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Parse()
}

var (
	service     = flag.String("service", "", "generate service name")
	templateUrl = flag.String("templateUrl", "git@github.com:WhiteMatrixTech/matrix-microservice-template.git", "from template")
	configPath  = flag.String("config", "config.yml", "config file path")
	createRepo  = flag.Bool("createRepo", false, "auto create repo to github")
)

func main() {
	var err error
	var c pkg.TemplateConfig
	config.DefaultConfig, err = config.NewConfig(
		config.WithEntity(&c),
		config.WithSource(
			file.NewSource(
				file.WithPath(*configPath))))
	if err != nil { // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err.Error())
	}
	if err = config.Scan(&c); err != nil {
		//if err = v.Unmarshal(&c); err != nil {
		log.Fatalln(err)
	}
	if *service != "" {
		c.Service = *service
	}
	if *templateUrl != "" {
		c.TemplateUrl = *templateUrl
	}
	if *createRepo {
		c.CreateRepo = *createRepo
	}
	if c.CreateRepo {
		if c.Github.Name == "" {
			c.Github.Name = c.Service
		}
	}
	if c.Destination == "" {
		c.Destination = "."
	}
	c.Destination = filepath.Join(c.Destination, c.Service)
	err = pkg.OldGenerate(&c)
	if err != nil {
		log.Fatalln(err)
	}
}
