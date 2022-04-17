/*
 * @Author: lwnmengjing
 * @Date: 2021/12/16 7:39 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/12/16 7:39 下午
 */

package pkg

import (
	"fmt"
)

type TemplateConfig struct {
	GenerationId  string        `yaml:"generationId"`
	Service       string        `yaml:"service"`
	TemplateUrl   string        `yaml:"templateUrl"`
	TemplateLocal string        `yaml:"templateLocal"`
	Destination   string        `yaml:"destination"`
	Github        *GithubConfig `yaml:"github"`
	Params        interface{}   `yaml:"params"`
	Ignore        []string      `yaml:"ignore"`
}

func (e *TemplateConfig) OnChange() {
	fmt.Println("config changed")
}
