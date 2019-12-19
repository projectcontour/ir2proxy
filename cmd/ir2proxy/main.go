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

func main() {

	exitcode := run()
	os.Exit(exitcode)

}

func run() int {

	log := logrus.StandardLogger()
	app := kingpin.New("ir2proxy", "Contour IngressRoute to HTTPProxy conversion tool.")

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
		fmt.Printf("---\n%s%s", outputWarnings, outputYAML)
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
