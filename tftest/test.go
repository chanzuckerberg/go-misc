package tftest

import (
	"errors"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
)

type TestMode int

const (
	Apply TestMode = 0
	Plan  TestMode = 1
	Init  TestMode = 2
)

// Test encapsulates and provides structure to a terratest-driven terraform test.
//
// Tests are run in 4 stages– Setup, Apply, Validate and Cleanup. Each stage will persist relevant
// data so that subsequent test runs can be isolated to a subset of stages.
//
// Setup Stage
//
// The setup Stage is used to create all the preconditions for running the terraform code under
// test. The user supplied Setup function must return a set of Options for running the code. In
// addition, it can create any additional resources that need to exist before running.
//
// Apply Stage
//
// The above options are used to initialize and apply the Terraform code under test. Note that Mode
// can be set to something other than Apply. The terraform state file is saved locally for use in
// Validate and Cleanup.
//
// Validate Stage
//
// If code was successfully applied, the user-supplied Validate function is run to make assertions
// about the resulting infrastructure. Note that if Mode is set to something other than Apply, the
// Validate function is not currently very useful.
//
// Cleanup Stage
//
// In addition to running a `terraform destroy`, and deleting any cached data (saved options,
// terraform state) the user-supplied Cleanup function is run to do arbitrary clean up work.
//
//
// Env Variables
//
// Each stage persists relevant data and can be skipped on subsequent runs. There are two
// environment variables which control which stages are run – SKIP and ONLY, each of which take a
// comma-separated list of stage names. Setting both is not allowed and will generate a test
// failure.
//
// Example–
//   SKIP=cleanup go test . -run TestFoo
// This will run the first three stages. If after that–
//  ONLY=validate go test . -run TestFoo
// ...the saved options from Setup and saved terraform state from Apply will be reused (and the
// infrastructure is presumably still up). This enables one to iterate quickly on testing terraform
// modules.
//
// Hopefully many useful workflows can be derived from these building blocks.
type Test struct {
	Setup    func(*testing.T) *terraform.Options
	Validate func(*testing.T, *terraform.Options)
	Cleanup  func(*testing.T, *terraform.Options)

	Mode        TestMode
	SkipDestroy bool

	skip []string
	only []string
}

func (tt *Test) validate() error {
	if tt.Setup == nil {
		return errors.New("Setup must be set")
	}

	if tt.Validate == nil {
		return errors.New("Validate must be set")
	}
	return nil
}

func (tt *Test) setupEnv(t *testing.T) {
	skip := ListEnvVar("SKIP")
	only := ListEnvVar("ONLY")

	if len(skip) > 0 && len(only) > 0 {
		t.Fatal("SKIP and ONLY env variables both set, you can only set one")
		return
	}
	tt.skip = skip
	tt.only = only
}

func (tt *Test) shouldRun(stage string) bool {
	if len(tt.only) > 0 {
		for _, s := range tt.only {
			if s == stage {
				return true
			}
		}
		return false
	}

	for _, s := range tt.skip {
		if s == stage {
			return false
		}
	}

	return true
}

func (tt *Test) Stage(t *testing.T, stage string, f func()) {
	if tt.shouldRun(stage) {
		t.Logf("running stage %s", stage)
		f()
	} else {
		t.Logf("skipping stage %s", stage)
	}
}

func (tt *Test) Run(t *testing.T) {
	r := require.New(t)

	terraformDirectory := "."

	err := tt.validate()
	r.NoError(err)

	tt.setupEnv(t)

	defer tt.Stage(t, "cleanup", func() {
		options := test_structure.LoadTerraformOptions(t, terraformDirectory)
		// for some tests we want to skip the destroy and let our GC processes clean up
		if !tt.SkipDestroy {
			terraform.DestroyE(t, options) //nolint
		}
		Clean(terraformDirectory)
		test_structure.CleanupTestDataFolder(t, terraformDirectory)
		if tt.Cleanup != nil {
			tt.Cleanup(t, options)
		}
	})

	tt.Stage(t, "setup", func() {
		fileExists := func(filename string) bool {
			info, err := os.Stat(filename)
			if os.IsNotExist(err) {
				return false
			}
			return !info.IsDir()
		}

		if fileExists(test_structure.FormatTestDataPath(terraformDirectory, "TerraformOptions.json")) {
			t.Log("options file exists, skipping generation")
			return
		}
		options := tt.Setup(t)
		test_structure.SaveTerraformOptions(t, terraformDirectory, options)
	})

	tt.Stage(t, "apply", func() {
		r := require.New(t)
		options := test_structure.LoadTerraformOptions(t, terraformDirectory)
		switch tt.Mode {
		case Apply:
			terraform.InitAndApply(t, options)
		case Plan:
			rc, err := terraform.InitAndPlanWithExitCodeE(t, options)
			r.NoError(err)
			r.Equal(2, rc)
		case Init:
			terraform.Init(t, options)
		}
	})

	tt.Stage(t, "validate", func() {
		options := test_structure.LoadTerraformOptions(t, terraformDirectory)
		tt.Validate(t, options)
	})
}
