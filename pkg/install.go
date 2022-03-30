package pkg

import (
	"fmt"
	"github.com/mss-boot-io/micro-service-gen-tool/version"
	"os"
	"strings"
)

// Install update generate-tool
func Install() {
	destPath := downloadLatest()
	//复制文件到对应目录
	copyStaticFile(destPath, "generate-tool")
	fmt.Println("Install completed")
}

// Upgrade check update
func Upgrade(ask bool) {
	if GetLatestVersionFromTagName() != version.Version {
		//need update
		fmt.Printf("do you need to upgrade [%s]: ", Yellow("y/n"))
		var upgrade bool
		if !ask {
			upgrade = true
		} else {
			var input string
			_, _ = fmt.Scan(&input)
			switch strings.ToLower(input) {
			case "y", "yes", "t", "true":
				upgrade = true
			}
		}
		if upgrade {
			Install()
			os.Exit(0)
		}
	}
}
