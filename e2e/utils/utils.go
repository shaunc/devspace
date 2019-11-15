package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/devspace-cloud/devspace/pkg/devspace/analyze"
	"github.com/devspace-cloud/devspace/pkg/devspace/kubectl"
	"github.com/devspace-cloud/devspace/pkg/devspace/services"
	"github.com/devspace-cloud/devspace/pkg/util/log"
)

// ChangeWorkingDir changes the working directory
func ChangeWorkingDir(wd string) error {
	// Change working directory
	err := os.Chdir(wd)
	if err != nil {
		return err
	}

	// Notify user that we are not using the current working directory
	log.Infof("Using devspace config in %s", filepath.ToSlash(wd))

	return nil
}

// PrintTestResult prints a test result with a specific formatting
func PrintTestResult(name string, err error) {
	successIcon := html.UnescapeString("&#" + strconv.Itoa(128513) + ";")
	failureIcon := html.UnescapeString("&#" + strconv.Itoa(128545) + ";")

	if err == nil {
		fmt.Printf("%v  %v successfully deployed!\n", successIcon, name)
	} else {
		log.Fatalf("%v  %v failed to deploy: %v\n", failureIcon, name, err)
	}
}

// DeleteNamespaceAndWait deletes a given namespace and waits for the process to finish
func DeleteNamespaceAndWait(client kubectl.Client, namespace string) {
	log.StartWait("Deleting namespace '" + namespace + "'")
	err := client.KubeClient().CoreV1().Namespaces().Delete(namespace, nil)
	if err != nil {
		log.Fatal(err)
	}

	isExists := true
	for isExists {
		_, err = client.KubeClient().CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if err != nil {
			isExists = false
		}
	}

	defer log.StopWait()
}

// AnalyzePods waits for the pods to be running (if possible) and healthcheck them
func AnalyzePods(client kubectl.Client, namespace string) error {
	var problems []string
	problems, err := analyze.Pods(client, namespace, false)
	bs, jsonErr := json.Marshal(problems)
	if jsonErr != nil {
		return jsonErr
	}
	if len(problems) > 0 {
		return errors.New("The following problems were found:" + string(bs))
	}
	if err != nil {
		return err
	}

	return nil
}

// PortForwardAndPing creates port-forwardings and ping them for a 200 status code
func PortForwardAndPing(servicesClient services.Client) error {
	portForwarder, err := servicesClient.StartPortForwarding()
	if err != nil {
		return err
	}

	for _, pf := range portForwarder {
		ports, err := pf.GetPorts()
		if err != nil {
			return err
		}

		for _, p := range ports {
			url := fmt.Sprintf("http://localhost:%v/", p.Local)
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}

			if resp.StatusCode == 200 {
				log.Donef("Pinging %v: status code 200", url)
			} else {
				return fmt.Errorf("Pinging %v: status code %v", url, resp.StatusCode)
			}
		}
	}

	// We close all the port-forwardings
	defer func() {
		for _, v := range portForwarder {
			v.Close()
		}
	}()

	return nil
}
