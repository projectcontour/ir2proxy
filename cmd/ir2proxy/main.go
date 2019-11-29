package main

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
	irv1beta1 "github.com/projectcontour/contour/apis/contour/v1beta1"
	contourscheme "github.com/projectcontour/contour/apis/generated/clientset/versioned/scheme"
	hpv1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	for _, yamldoc := range bytes.Split(data, []byte("---")) {
		if len(yamldoc) == 0 {
			continue
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode

		ir, groupVersionKind, err := decode(yamldoc, nil, nil)
		if err != nil {
			log.Infof("Skipped yaml doc, %s", err)
			continue
		}
		switch t := ir.(type) {
		case *irv1beta1.IngressRoute:
			log.Infof("IngressRoute %s, namespace %s, labels %s", t.ObjectMeta.Name, t.ObjectMeta.Namespace, t.ObjectMeta.Labels)
			hp, err := translateIngressRouteToHTTPProxy(t)
			if err != nil {
				log.Fatal(err)
			}
			log.Error(spew.Sdump(hp))
		default:
			log.Errorf("This utility only works with IngressRoute, a %s was supplied.", groupVersionKind)
		}

	}

}

func verifyYAMLFile(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()

}

func translateIngressRouteToHTTPProxy(ir *irv1beta1.IngressRoute) (hp *hpv1.HTTPProxy, err error) {

	hp = &hpv1.HTTPProxy{
		ObjectMeta: v1.ObjectMeta{
			Name:        ir.ObjectMeta.Name,
			Namespace:   ir.ObjectMeta.Namespace,
			Labels:      ir.ObjectMeta.DeepCopy().GetLabels(),
			Annotations: ir.ObjectMeta.DeepCopy().GetAnnotations(),
		},
	}

	if ir.Spec.VirtualHost != nil {
		hp.Spec.VirtualHost = ir.Spec.VirtualHost
	}

	return hp, err
}
