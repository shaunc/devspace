package main

import (
	"flag"
	"fmt"
	"github.com/devspace-cloud/devspace/e2e/tests/examples"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"os"
	"strings"
)

var testNamespace = "examples-test-namespace"

var tests = map[string]*[]string{
	"enter":    &[]string{},
	"deploy":   &[]string{"default", "profile", "kubectl", "helm"},
	"init":     &[]string{},
	"examples": &[]string{"quickstart", "kustomize", "profiles", "microservices", "minikube", "quickstart-kubectl", "php-mysql", "dependencies"},
}

// Create a new type for a list of Strings
type stringList []string

// Implement the flag.Value interface
func (s *stringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringList) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

type Test func(subTests []string, ns string, pwd string) error

var availableTests = map[string]Test{
	"examples": examples.Run,
}

var subTests = map[string]*stringList{}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testCommand := flag.NewFlagSet("test", flag.ExitOnError)
	listCommand := flag.NewFlagSet("list", flag.ExitOnError)

	for t := range availableTests {
		subTests[t] = &stringList{}
		testCommand.Var(subTests[t], "test-"+t, "A comma seperated list of sub tests to be passed")
	}

	var test stringList
	testCommand.Var(&test, "test", "A comma seperated list of group tests to pass")

	var testlist stringList
	listCommand.Var(&testlist, "test", "A comma seperated list of group tests to list (leave empty to list all group tests)")

	// Verify that a subcommand has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if len(os.Args) < 2 {
		fmt.Println("test or list subcommand is required")
		os.Exit(1)
	}

	// Switch on the subcommand
	// Parse the flags for appropriate FlagSet
	// FlagSet.Parse() requires a set of arguments to parse as input
	// os.Args[2:] will be all arguments starting after the subcommand at os.Args[1]
	switch os.Args[1] {
	case "list":
		listCommand.Parse(os.Args[2:])
	case "test":
		testCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	// FlagSet.Parse() will evaluate to false if no flags were parsed (i.e. the user did not provide any flags)
	// If "list" and "test" are used together, only the former will be parsed and recognized, the latter will be ignored
	if listCommand.Parsed() {
		// Required Flags
		fmt.Println("listCommand parsed!")
	}
	if testCommand.Parsed() {
		// We gather all the group tests called with the --test flag. e.g: --test=examples,init
		var testsToRun = map[string]Test{}
		for _, testName := range test {
			if tests[testName] == nil {
				// arg is not valid
				fmt.Printf("'%v' is not a valid argument for --test. Valid arguments are the following: [ ", testName)
				for key := range tests {
					fmt.Printf("%v ", key)
				}
				fmt.Printf("]\n ")
				os.Exit(1)
			}
			testsToRun[testName] = availableTests[testName]
		}

		// If cmd test alone (if no --test flag), we want to run all available tests
		if len(testsToRun) == 0 {
			for testName := range availableTests {
				testsToRun[testName] = availableTests[testName]
			}
		}

		for testName, testRun := range testsToRun {
			parameterSubTests := []string{}
			if t, ok := subTests[testName]; ok && t != nil && len(*t) > 0 {
				for _, s := range *t {
					if !utils.StringInSlice(s, *tests[testName]) {
						// arg is not valid
						fmt.Printf("'%v' is not a valid argument for --test-%v. Valid arguments are the following: [ ", s, testName)
						for _, st := range *tests[testName] {
							fmt.Printf("%v ", st)
						}
						fmt.Printf("]\n ")
						os.Exit(1)
					}
					parameterSubTests = append(parameterSubTests, s)
				}
			}

			// We run the actual group tests by passing the sub tests
			err := testRun(parameterSubTests, testNamespace, pwd)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	fmt.Println("End of parsing!")

	//examples.TestDeploy(testNamespace, pwd)
	//init.TestInit(pwd)
	//enter.TestEnter(pwd, testNamespace)
}

// go run . test --test=bla --test-examples --test-enter
// go run . list --test=examples

/*
	flags.StringVarP(&globalFlags.Profile, "profile", "p", "", "The devspace profile to use (if there is any)")
	flags.StringSliceVar(&globalFlags.Vars, "var", []string{}, "Variables to override during execution (e.g. --var=MYVAR=MYVALUE)")

	deployCmd.Flags().BoolVarP(&cmd.ForceBuild, "force-build", "b", false, "Forces to (re-)build every image")
	deployCmd.Flags().BoolVarP(&cmd.ForceDeploy, "force-examples", "d", false, "Forces to (re-)examples every deployment")
	deployCmd.Flags().BoolVar(&cmd.ForceDependencies, "force-dependencies", false, "Forces to re-evaluate dependencies (use with --force-build --force-examples to actually force building & deployment of dependencies)")
	deployCmd.Flags().StringVar(&cmd.Deployments, "deployments", "", "Only examples a specifc deployment (You can specify multiple deployments comma-separated")

Test 1 - default
1. examples (without profile & var)
2. examples --force-build & check if rebuild
3. examples --force-examples & check NO build but deployed
4. examples --force-dependencies & check NO build & check NO deployment but dependencies are deployed
5. examples --force-examples --deployments=test1,test2 & check NO build & only deployments deployed

Test 2 - profile
1. examples --profile=bla --var var1=two --var var2=three
2. examples --profile=bla --var var1=two --var var2=three --force-build & check if rebuild
3. examples --profile=bla --var var1=two --var var2=three --force-examples & check NO build but deployed
4. examples --profile=bla --var var1=two --var var2=three --force-dependencies & check NO build & check NO deployment but dependencies are deployed
4. examples --profile=bla --var var1=two --var var2=three --force-examples --deployments=test1,test2 & check NO build & only deployments deployed

Test 3 - kubectl
1. examples & kubectl (see quickstart-kubectl)
2. purge (check if everything is deleted except namespace)

Test 4 - helm
1. examples & helm (see quickstart) (v1beta5 no tiller)
2. purge (check if everything is deleted except namespace)

Test 1
	enterCmd.Flags().StringVarP(&cmd.Container, "container", "c", "", "Container name within pod where to execute command")
	enterCmd.Flags().StringVar(&cmd.Pod, "pod", "", "Pod to open a shell to")
	enterCmd.Flags().StringVarP(&cmd.LabelSelector, "label-selector", "l", "", "Comma separated key=value selector list (e.g. release=test)")
	enterCmd.Flags().BoolVar(&cmd.Pick, "pick", false, "Select a pod")

1. enter --container
2. enter --pod
3. enter --label-selector
4. enter --pick

*/

// enter test with label selector
