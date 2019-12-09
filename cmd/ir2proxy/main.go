package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

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
	app := kingpin.New("ir2proxy", "Contour IngressRoute to HTTPPRoxy conversion tool.")

	yamlfile := app.Arg("yaml", "YAML file to parse for IngressRoute objects").Required().ExistingFile()

	args := os.Args[1:]
	kingpin.MustParse(app.Parse(args))

	data, err := ioutil.ReadFile(*yamlfile)
	if err != nil {
		log.Error(err)
		return 1
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

		hp, translationErrors := translator.IngressRouteToHTTPProxy(ir)
		if len(translationErrors) > 0 {
			if hp == nil {
				// If we didn't get a HTTPProxy back, then there was at least
				// one fatal error. Log them and exit.
				for _, translationError := range translationErrors {
					log.Error(translationError)
				}
				return 1
			}
			for _, translationError := range translationErrors {
				log.Warn(translationError)
			}
		}

		outputYAML, err := yaml.Marshal(hp)
		if err != nil {
			log.Warn(err)
			return 1
		}
		fmt.Printf("---\n%s", outputYAML)
	}

	return 0
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
