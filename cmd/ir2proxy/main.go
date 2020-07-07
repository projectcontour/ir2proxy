// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/projectcontour/ir2proxy/internal/k8sdecoder"
	"github.com/projectcontour/ir2proxy/internal/translator"
	"github.com/projectcontour/ir2proxy/internal/validate"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var build = "devel"

func main() {

	exitcode := run()
	os.Exit(exitcode)

}

func run() int {

	log := logrus.StandardLogger()
	app := kingpin.New("ir2proxy", "Contour IngressRoute to HTTPProxy conversion tool.")
	app.Version(build)
	yamlfile := app.Arg("yaml", "YAML file to parse for IngressRoute objects").Required().ExistingFile()

	args := os.Args[1:]
	kingpin.MustParse(app.Parse(args))

	data, err := ioutil.ReadFile(*yamlfile)
	if err != nil {
		log.Error(err)
	}

	for _, yamldoc := range splitYAML(data) {
		ir, err := k8sdecoder.DecodeIngressRoute(yamldoc)
		if err != nil {
			log.Error(err)
			return 1
		}

		validationErrors := validate.CheckIngressRoute(ir)
		if len(validationErrors) > 0 {
			for _, validationError := range validationErrors {
				log.Error(validationError)
			}
			return 1
		}

		hp, warnings, err := translator.IngressRouteToHTTPProxy(ir)
		if err != nil {
			log.Error(err)
			return 1
		}
		for _, warning := range warnings {
			log.Warn(warning)
		}

		outputYAML, err := yaml.Marshal(hp)
		if err != nil {
			log.Warn(err)
			return 1
		}
		// The Kubernetes standard header field `currentTimestamp` serializes weirdly,
		// so filter it out.
		// See https://github.com/projectcontour/ir2proxy/issues/8 for more explanation here.
		outputYAML = bytes.ReplaceAll(outputYAML, []byte("  creationTimestamp: null\n"), []byte(""))
		outputWarnings := commentedWarnings(warnings)
		fmt.Printf("---\n%s\n%s", outputWarnings, outputYAML)
	}

	return 0
}

func commentedWarnings(warnings []string) string {
	for index, warning := range warnings {
		warnings[index] = "# " + strings.ReplaceAll(warning, ". ", ".\n# ")
	}
	return strings.Join(warnings, "\n")

}

func splitYAML(yamldata []byte) [][]byte {

	var yamldocs [][]byte
	for _, yamldoc := range bytes.Split(yamldata, []byte("---")) {
		if len(yamldoc) == 0 {
			continue
		}
		yamldocs = append(yamldocs, yamldoc)
	}
	return yamldocs

}
