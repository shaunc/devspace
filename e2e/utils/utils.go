package utils

import (
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/devspace-cloud/devspace/pkg/devspace/analyze"
	"github.com/devspace-cloud/devspace/pkg/devspace/kubectl"
	"github.com/devspace-cloud/devspace/pkg/devspace/services"
	logger "github.com/devspace-cloud/devspace/pkg/util/log"
)

// ChangeWorkingDir changes the working directory
func ChangeWorkingDir(pwd string) error {
	log := logger.GetInstance()

	wd, err := filepath.Abs(pwd)
	if err != nil {
		return err
	}
	fmt.Println("WD:", wd)
	// Change working directory
	err = os.Chdir(wd)
	if err != nil {
		return err
	}

	// Notify user that we are not using the current working directory
	log.Infof("Using devspace config in %s", filepath.ToSlash(wd))

	return nil
}

// PrintTestResult prints a test result with a specific formatting
func PrintTestResult(testName string, subTestName string, err error) {
	successIcon := html.UnescapeString("&#" + strconv.Itoa(128513) + ";")
	failureIcon := html.UnescapeString("&#" + strconv.Itoa(128545) + ";")

	if err == nil {
		fmt.Printf("%v  Test '%v' of group test '%v' successfully passed!\n", successIcon, subTestName, testName)
	} else {
		fmt.Printf("%v  Test '%v' of group test '%v' failed!\n", failureIcon, subTestName, testName)
	}
}

// DeleteNamespaceAndWait deletes a given namespace and waits for the process to finish
func DeleteNamespaceAndWait(client kubectl.Client, namespace string) {
	log := logger.GetInstance()

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
	err := analyze.NewAnalyzer(client, logger.GetInstance()).Analyze(namespace, false)
	if err != nil {
		return err
	}

	return nil
}

// PortForwardAndPing creates port-forwardings and ping them for a 200 status code
func PortForwardAndPing(servicesClient services.Client) error {
	log := logger.GetInstance()

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
				return fmt.Errorf("pinging %v: status code %v", url, resp.StatusCode)
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

// Equal tells whether a and b contain the same elements.
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

/* The MIT License (MIT)

Copyright (c) 2018 otiai10

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	return copy(src, dest, info)
}

// copy dispatches copy-funcs according to the mode.
// Because this "copy" could be called recursively,
// "info" MUST be given here, NOT nil.
func copy(src, dest string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return lcopy(src, dest, info)
	}
	if info.IsDir() {
		return dcopy(src, dest, info)
	}
	return fcopy(src, dest, info)
}

// fcopy is for just a file,
// with considering existence of parent directory
// and file permission.
func fcopy(src, dest string, info os.FileInfo) error {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

// dcopy is for a directory,
// with scanning contents inside the directory
// and pass everything to "copy" recursively.
func dcopy(srcdir, destdir string, info os.FileInfo) error {

	if err := os.MkdirAll(destdir, info.Mode()); err != nil {
		return err
	}

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := copy(cs, cd, content); err != nil {
			// If any error, exit immediately
			return err
		}
	}
	return nil
}

// lcopy is for a symlink,
// with just creating a new symlink by replicating src symlink.
func lcopy(src, dest string, info os.FileInfo) error {
	src, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(src, dest)
}

// =====================================================================

// CreateTempDir creates a temp directory in /tmp
func CreateTempDir() (dirPath string, dirName string, err error) {
	// Create temp dir in /tmp/
	dirPath, err = ioutil.TempDir("", "init")
	dirName = filepath.Base(dirPath)
	if err != nil {
		return
	}
	fmt.Println("tempDir created:", dirPath)
	return
}

// DeleteTempDir deletes temp directory
func DeleteTempDir(dirPath string) {
	log := logger.GetInstance()

	//Delete temp folder
	err := os.RemoveAll(dirPath)
	if err != nil {
		log.Fatalf("Error removing dir: %v", err)
	}
}

// Capture replaces os.Stdout with a writer that buffers any data written
// to os.Stdout. Call the returned function to cleanup and get the data
// as a string.
func Capture() func() (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	done := make(chan error, 1)

	save := os.Stdout
	os.Stdout = w

	var buf strings.Builder

	go func() {
		_, err := io.Copy(&buf, r)
		r.Close()
		done <- err
	}()

	return func() (string, error) {
		os.Stdout = save
		w.Close()
		err := <-done
		return buf.String(), err
	}
}

// StringInSlice checks if a string is in a slice
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// DeleteTempAndResetWorkingDir deletes /tmp dir and reinitialize the working dir
func DeleteTempAndResetWorkingDir(tmpDir string, pwd string) {
	DeleteTempDir(tmpDir)
	_ = ChangeWorkingDir(pwd)
}

// LookForDeployment search for a specific deployment name among the deployments, returns true if found
func LookForDeployment(client kubectl.Client, namespace string, expectedDeployment ...string) (bool, error) {
	s, err := client.KubeClient().CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	var deployments []string

	for _, x := range s.Items {
		deployments = append(deployments, x.Name)
	}

	for _, d := range expectedDeployment {
		exists := StringInSlice(d, deployments)
		if !exists {
			return false, nil
		}
	}

	return true, nil
}
