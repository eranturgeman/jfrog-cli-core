package jas

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/xray/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/xray/services"
	xrayUtils "github.com/jfrog/jfrog-client-go/xray/services/utils"
	"os"
)

var (
	analyzerManagerExecuter utils.AnalyzerManagerInterface = &utils.AnalyzerManager{}
	skippedDirs                                            = []string{"**/*test*/**", "**/*venv*/**", "**/*node_modules*/**", "**/*target*/**"}
)

func GetExtendedScanResults(xrayResults []services.ScanResponse, dependencyTrees []*xrayUtils.GraphNode,
	serverDetails *config.ServerDetails, scannedTechnologies []coreutils.Technology) (*utils.ExtendedScanResults, error) {
	if serverDetails == nil || len(serverDetails.Url) == 0 {
		log.Warn("To include 'Advanced Security' scan as part of the audit output, please run the 'jf c add' command before running this command.")
		return &utils.ExtendedScanResults{XrayResults: xrayResults}, nil
	}
	analyzerManagerExist, err := analyzerManagerExecuter.ExistLocally()
	if err != nil {
		return &utils.ExtendedScanResults{XrayResults: xrayResults}, err
	}
	if !analyzerManagerExist {
		log.Debug("Since the 'Analyzer Manager' doesn't exist locally, its execution is skipped.")
		return &utils.ExtendedScanResults{XrayResults: xrayResults}, nil
	}
	applicabilityScanResults, eligibleForApplicabilityScan, err := getApplicabilityScanResults(xrayResults,
		dependencyTrees, serverDetails, scannedTechnologies, analyzerManagerExecuter)
	if err != nil {
		return nil, err
	}
	secretsScanResults, eligibleForSecretsScan, err := getSecretsScanResults(serverDetails, analyzerManagerExecuter)
	if err != nil {
		return nil, err
	}
	iacScanResults, eligibleForIacScan, err := getIacScanResults(serverDetails, analyzerManagerExecuter)
	if err != nil {
		return nil, err
	}
	return &utils.ExtendedScanResults{
		EntitledForJas:               true,
		XrayResults:                  xrayResults,
		ScannedTechnologies:          scannedTechnologies,
		ApplicabilityScanResults:     applicabilityScanResults,
		EligibleForApplicabilityScan: eligibleForApplicabilityScan,
		SecretsScanResults:           secretsScanResults,
		EligibleForSecretScan:        eligibleForSecretsScan,
		IacScanResults:               iacScanResults,
		EligibleForIacScan:           eligibleForIacScan,
	}, nil
}

func deleteJasProcessFiles(configFile string, resultFile string) error {
	exist, err := fileutils.IsFileExists(configFile, false)
	if err != nil {
		return err
	}
	if exist {
		if err = os.Remove(configFile); err != nil {
			return errorutils.CheckError(err)
		}
	}
	exist, err = fileutils.IsFileExists(resultFile, false)
	if err != nil {
		return err
	}
	if exist {
		err = os.Remove(resultFile)
	}
	return errorutils.CheckError(err)
}
