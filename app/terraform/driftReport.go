package terraform

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	tail "github.com/hpcloud/tail"

	"tfdrift/log"
)

// Parse out.tfplan and return the last line if it contains "Plan".
func GetResourceModificationCount(planFileRawString string, projectName string) (string, error) {
	var filename string = fmt.Sprintf("/tmp/%s-tmp", projectName)
	file, err := os.Create(filename)
	if err != nil {
		log.Infof("%s", err)
		panic(err)
	} else {
		file.WriteString(planFileRawString)
	}
	file.Close()
	log.Debugf("[GetResourceModificationCount] Opening file: %s", filename)

	var tailLine string = ""
	t, err := tail.TailFile(filename, tail.Config{Follow: true})
	for line := range t.Lines {
		noChanges, _ := regexp.MatchString("\\bNo changes\\b", line.Text)
		matched, err := regexp.MatchString("\\bPlan\\b", line.Text)
		if err != nil {
			panic(err)
		} else if matched {
			tailLine = line.Text
			return tailLine, err
		} else if noChanges {
			tailLine = "Plan: 0 to add, 0 to change, 0 to destroy."
			return tailLine, nil
		}
	}

	return tailLine, err
}

// Get # of Add, Change, Destroy in out.tfplan
func ParseResourceModificationCount(resourceModificationString string) (map[string]int, error) {
	var resourceModification map[string]int
	resourceModification = make(map[string]int)

	re := regexp.MustCompile("[0-9]+")
	var counts []string = re.FindAllString(resourceModificationString, -1)

	countAdd, err := strconv.Atoi(counts[0])
	if err != nil {
		panic(err)
	}
	countChange, err := strconv.Atoi(counts[1])
	if err != nil {
		panic(err)
	}
	countDestroy, err := strconv.Atoi(counts[2])
	if err != nil {
		panic(err)
	}

	resourceModification["CountAdd"] = countAdd
	resourceModification["CountChange"] = countChange
	resourceModification["CountDestroy"] = countDestroy

	return resourceModification, err
}

// terraform plan -detailed-exitcode
// 0 = false (no changes)
// 1 = Error
// 2 = true  (drift)
func GetDriftSummary(exitCode int, planErr error, state *tfjson.State, project string) string {
	var message string
	log.Debugf("[GetDriftSummary] EXITCODE %d", exitCode)
	if exitCode == 2 {
		message = "Drift detected for Plan."
		log.Debugf("[GetDriftSummary] %s", message)
	} else if exitCode == 0 {
		message = "No changes."
		log.Debugf("[GetDriftSummary] %s", message)
	} else if exitCode == 1 {
		message = "Failed to run tfxec on project: " + project
		log.Debugf("[GetDriftSummary] %s", message)
	} else {
		message = fmt.Sprintf("Improper exit code of %s returned.", exitCode)
		log.Debugf("[GetDriftSummary] %s", message)
	}
	return message
}

// Populate a TerraformService structure with relevant data.
func UpdateDriftReportData(state *tfjson.State, projectName string, counts map[string]int, summary string) *TerraformService {
	tfs := &TerraformService{
		//State:            state,
		ProjectName:      projectName,
		TerraformVersion: state.TerraformVersion,
		CountAdd:         counts["CountAdd"],
		CountChange:      counts["CountChange"],
		CountDestroy:     counts["CountDestroy"],
		Summary:          summary,
	}
	return tfs
}

// The function that actually counts the most.
func DriftReport(absProjectPath string, backendConfig string, terraformVersion string) *TerraformService {
	// Pre-Init
	CleanupCachedFiles(absProjectPath)

	// tfexec Setup
	service := ConfigureTerraform(absProjectPath, terraformVersion)
	_, projectName := GetProjectName(absProjectPath)
	// terraform init
	project, failedProject, err := Init(service, backendConfig)
	if err != nil {
		log.Infof("[DriftReport] Failed project: %s", project)
	}

	var tfService *TerraformService = &TerraformService{}

	if !failedProject {
		// terraform show
		state := Show(service)
		// terraform plan (-detailed-exitcode)
		exitCode, planErr := Plan(service, projectName)

		// terraform plan (-out=out.tfplan)
		planPath := fmt.Sprintf("%s/%s.tfplan", absProjectPath, projectName)

		var rawPlan, showPlanErr = ShowPlanFileRaw(service, planPath)
		log.Debugf("[DriftReport] Retrieved rawPlan for project: %s", project)

		// If no rawPlan is able to be found, skip this and set ResourceMod count to 0,0,0 :'(
		var resourceCount map[string]int = map[string]int{"CountAdd": 0, "CountChange": 0, "CountDestroy": 0}
		if rawPlan != "" {
			planString, err := GetResourceModificationCount(rawPlan, projectName)
			if err != nil {
				panic(err)
			}
			log.Infof("[DriftReport] Retrieved resource count for project: %s.", project)

			// If drift detected in Plan return the Add/Change/Destroy count values.
			modifiedResourceCount, err := ParseResourceModificationCount(planString)
			if err != nil {
				panic(err)
			}
			log.Infof("[DriftReport] Parsing rawPlan to get modified resource count for project: %s.", project)
			resourceCount = modifiedResourceCount
		}

		// Determine error
		var terraformError error
		if planErr == nil && showPlanErr != nil {
			terraformError = showPlanErr
			log.Debugf("[DriftReport] Error for %s is %s", project, terraformError)
		} else {
			terraformError = planErr
			log.Debugf("[DriftReport] Error for %s is %s", project, terraformError)
		}

		// Get project name + status information
		summary := GetDriftSummary(exitCode, terraformError, state, project)
		log.Debugf("[DriftReport] Getting Drift Summary for %s", project)

		// Format a TerraformService structure with all information needed for the Drift Report
		tfService := UpdateDriftReportData(state, projectName, resourceCount, summary)
		log.Debugf("[DriftReport] Returning Terraform Service for %s", project)
		return tfService
	}

	return tfService
}

func GeneratePlan(absProjectPath string, backendConfig string, terraformVersion string) *TerraformService {
	service := ConfigureTerraform(absProjectPath, terraformVersion)
	// terraform init
	var tfService *TerraformService = &TerraformService{}

	_, projectName := GetProjectName(absProjectPath)
	project, failedProject, err := Init(service, backendConfig)
	if err != nil {
		log.Infof("[DriftReport] Failed project: %s", project)
	}
	log.Infof("Generating plan for %s", projectName)

	if !failedProject {
		// terraform show
		Show(service)
		// terraform plan (-detailed-exitcode)
		_, planErr := Plan(service, projectName)

		var terraformError error
		if planErr != nil {
			terraformError = planErr
			log.Debugf("[DriftReport] Error for %s is %s", project, terraformError)
		}
		return tfService
	}
	log.Infof("Finished Generating Plan")
	return tfService
}

func TerraformPlanTrim(s string) string {
	if idx := strings.Index(s, "Terraform will perform the following actions"); idx != -1 {
		return s[idx:]
	}
	return s
}

// Go channel which returns the result of a DriftReport (required to parallelize)
func GetProjectDrift(ch chan *TerraformService, absProjectPath string, backendConfig string, terraformVersion string) {
	log.Printf("[GetDriftReport] Getting values for project: %s", absProjectPath)
	ch <- DriftReport(absProjectPath, backendConfig, terraformVersion)
}

// func GetDiff(ch chan *TerraformService, absProjectPath string, providerName string, backendConfig string) {
// 	log.Printf("[GetDriftDiff] Getting values for project: %s", absProjectPath)
// 	ch <- GetPlanDiff(absProjectPath, providerName, backendConfig)
// }

func GetPlan(ch chan *TerraformService, absProjectPath string, backendConfig string, terraformVersion string) {
	log.Printf("[GetDriftDiff] Getting values for project: %s", absProjectPath)
	ch <- GeneratePlan(absProjectPath, backendConfig, terraformVersion)
}
