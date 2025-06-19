package main

import (
	"context"
	"tfdrift/app/general"
	"tfdrift/app/terraform"
	"tfdrift/log"
	"time"

	"github.com/spf13/cobra"
)

var (
	// This gets set during the compilation. See below.
	path         string
	html         bool
	backendConfig string
	terraformVersion string
)

var TerraformContext = context.Background()

func main() {
	var scan = &cobra.Command{
		Use:   "scan",
		Short: "scan for infrastructure drift",
		Run: func(cmd *cobra.Command, args []string) {
			optionOutput := ""
			cvProjects, cvIsPlannable := general.GetPlannableProjects(path)

			driftDetectTime := time.Now()
			var terraformServices []*terraform.TerraformService
			tfChannel := make(chan *terraform.TerraformService)

			//tf drift is adding a batching job because too many concurrencies during terraform init will block some of them from running
			// https://github.com/hashicorp/terraform/issues/32915
			if cvIsPlannable {
				batchSize := 5 // Number of goroutines per batch
				for i := 0; i < len(cvProjects); i += batchSize {
					for j := i; j < i+batchSize && j < len(cvProjects); j++ {
						go func(absProjectPath string) {
							terraform.GetProjectDrift(tfChannel, absProjectPath, backendConfig, terraformVersion)
						}(cvProjects[j])
					}
					time.Sleep(5 * time.Second) // Wait for 2 seconds before starting the next batch
				}
				for _, _ = range cvProjects {
					terraformServices = append(terraformServices, <-tfChannel)
				}
			} else {
				log.Printf("[reportCmd] No *.tf files found")
			}
			// Where is the message going?
			if optionOutput == "stdout" || optionOutput == "" {
				log.Debug("[cmdReport] Outputting to Stdout.")
				if html {
					terraform.GenerateHTML(terraformServices)
				}
				terraform.PrettyTable(terraformServices)

				// Drift Report
				log.Printf("[reportCmd] Drift report took %s to report to stdout.\n", time.Since(driftDetectTime))
			} else {
				log.Errorf("[cmdReport] optionOutput: [%s] not supported (discord, stdout)", optionOutput)
			}
		},
	}

	var rootCmd = &cobra.Command{Use: "tfdrift"}
	scan.Flags().StringVar(&path, "path", "", "path to scan")
	scan.Flags().BoolVar(&html, "html", false, "path to scan")
	scan.Flags().StringVar(&backendConfig, "backend-config", "", "backend configuration file")
	scan.Flags().StringVar(&terraformVersion, "terraform-version", "1.7.0", "terraform version to use")

	rootCmd.AddCommand(scan)
	rootCmd.Execute()
}
