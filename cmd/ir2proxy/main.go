package main

import (
	"io/ioutil"
	"os"

	irv1beta1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	contourscheme "github.com/projectcontour/contour/apis/generated/clientset/versioned/scheme"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	log := logrus.StandardLogger()
	app := kingpin.New("ir2proxy", "Contour IngressRoute to HTTPPRoxy conversion tool.")

	yamlfile := app.Arg("yaml", "YAML file to parse for IngressRoute objects").Required().String()

	args := os.Args[1:]
	//kingpin.MustParse(app.Parse(args))
	app.Parse(args)

	if *yamlfile == "" {
		app.FatalUsage("Need a YAML file to operate on.")
	}

	if !verifyYAMLFile(*yamlfile) {
		log.Fatalf("File %s does not exist", *yamlfile)
	}

	data, err := ioutil.ReadFile(*yamlfile)
	if err != nil {
		log.Fatal(err)
	}

	contourscheme.AddToScheme(scheme.Scheme)

	decode := scheme.Codecs.UniversalDeserializer().Decode

	ir, groupVersionKind, err := decode(data, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	switch t := ir.(type) {
	case *irv1beta1.IngressRoute:
		log.Info("This was an IngressRoute")
		//log.Infof("%#v", t)
		log.Infof("IngressRoute %s, namespace %s", t.ObjectMeta.Name, t.ObjectMeta.Namespace)
	default:
		log.Infof("This utility only works with IngressRoute, a %s was supplied.", groupVersionKind)
	}

}

func verifyYAMLFile(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()

}
