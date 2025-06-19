package terraform

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	tfexec "github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"

	"tfdrift/log"
)

var TerraformContext = context.Background()

func ConfigureTerraform(workingDir string, terraformVersion string) *tfexec.Terraform {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(terraformVersion)),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}
	return tf
}

// Run `terraform init` so that the working directories context can be initialized.
func Init(tf *tfexec.Terraform, backendConfig ...string) (string, bool, error) {
	var project string = tf.WorkingDir()
	var failed bool = false

	var initOptions []tfexec.InitOption
	initOptions = append(initOptions, tfexec.Upgrade(true))

	if len(backendConfig) > 0 && backendConfig[0] != "" {
		initOptions = append(initOptions, tfexec.BackendConfig(backendConfig[0]))
	}

	err := tf.Init(TerraformContext, initOptions...)
	if err != nil {
		fmt.Println(err)
		failed = true
	}
	return project, failed, err
}

// (-detailed-exitcode)
// Run `terraform plan` against the state defined in the working directory.
// 0 = false (no changes)
// 1 = Error
// 2 = true  (drift)
func Plan(tf *tfexec.Terraform, outputName string) (int, error) {
	var exitCode int

	isPlanned, err := tf.Plan(TerraformContext, tfexec.Out(fmt.Sprintf("%s.tfplan", outputName)))
	if err != nil {
		exitCode = 1
		return exitCode, err
	}
	if isPlanned {
		exitCode = 2
	} else {
		exitCode = 0
	}

	return exitCode, err
}

// View State after it's been initialized and refreshed
// Run `terraform show` against the state defined in the working directory.
func Show(tf *tfexec.Terraform) *tfjson.State {
	state, err := tf.Show(TerraformContext)
	if err != nil {
		panic(err)
	}
	return state
}

// Run `terraform plan` against the state defined in the working directory.
func ShowPlanFileRaw(tf *tfexec.Terraform, planPath string) (string, error) {
	plan, err := tf.ShowPlanFileRaw(TerraformContext, planPath)
	if err != nil {
		return "", err
	}
	return plan, err
}
