package testdeploy

// RunQuickstart runs the test for the quickstart example
func RunDependencies(f *customFactory) error {
	f.GetLog().Info("Run Dependencies")

	err := RunTest(f, "dependencies", nil)
	if err != nil {
		return err
	}

	return nil
}
