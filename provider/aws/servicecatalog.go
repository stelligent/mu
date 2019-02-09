package aws

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/aws/aws-sdk-go/service/servicecatalog/servicecatalogiface"
	"github.com/stelligent/mu/common"
)

type serviceCatalogManager struct {
	dryrun       bool
	scAPI        servicecatalogiface.ServiceCatalogAPI
	stackManager common.StackManager
}

// newCatalogManager creates a new CatalogManager backed by service catalog
func newCatalogManager(sess *session.Session, dryrun bool, stackManager common.StackManager) (common.CatalogManager, error) {
	if dryrun {
		log.Debugf("Running in DRYRUN mode")
	}
	log.Debug("Connecting to Service Catalog service")
	scAPI := servicecatalog.New(sess)

	return &serviceCatalogManager{
		dryrun:       dryrun,
		scAPI:        scAPI,
		stackManager: stackManager,
	}, nil

}

func (scManager *serviceCatalogManager) GetStackID(recordID string) (string, error) {
	record, err := scManager.scAPI.DescribeRecord(&servicecatalog.DescribeRecordInput{
		Id: aws.String(recordID),
	})
	if err != nil {
		return "", err
	}
	for _, ro := range record.RecordOutputs {
		if aws.StringValue(ro.OutputKey) == "CloudformationStackARN" {
			return aws.StringValue(ro.OutputValue), nil
		}
	}

	return "", fmt.Errorf("Unable to find stack ARN for record '%s'", recordID)
}

func (scManager *serviceCatalogManager) createProvisionedProduct(productID string, artifactID string, name string, params map[string]string) error {
	// create new
	if scManager.dryrun {
		log.Infof("  DRYRUN: Skipping creation of product '%s' from artifact '%s'", productID, artifactID)
		return nil
	}
	provisioningParameters := make([]*servicecatalog.ProvisioningParameter, 0)
	for key, value := range params {
		provisioningParameters = append(provisioningParameters, &servicecatalog.ProvisioningParameter{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	ppOut, err := scManager.scAPI.ProvisionProduct(&servicecatalog.ProvisionProductInput{
		ProductId:              aws.String(productID),
		ProvisioningArtifactId: aws.String(artifactID),
		ProvisionedProductName: aws.String(name),
		ProvisioningParameters: provisioningParameters,
		ProvisionToken:         aws.String(strconv.FormatInt(time.Now().Unix(), 16)),
	})
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 5)
	stackID, err := scManager.GetStackID(aws.StringValue(ppOut.RecordDetail.RecordId))
	if err != nil {
		return err
	}

	log.Infof("  Creating stack '%s'", stackID)
	scManager.stackManager.AwaitFinalStatus(strings.Split(stackID, "/")[1])
	return nil
}

func (scManager *serviceCatalogManager) updateProvisionedProduct(productID string, artifactID string, name string, params map[string]string) error {
	// update existing
	if scManager.dryrun {
		log.Infof("  DRYRUN: Skipping update of product '%s' from artifact '%s'", productID, artifactID)
		return nil
	}
	provisioningParameters := make([]*servicecatalog.UpdateProvisioningParameter, 0)
	for key, value := range params {
		param := &servicecatalog.UpdateProvisioningParameter{
			Key: aws.String(key),
		}
		if value == "" {
			param.UsePreviousValue = aws.Bool(true)
		} else {
			param.Value = aws.String(value)
		}
		provisioningParameters = append(provisioningParameters, param)
	}
	ppOut, err := scManager.scAPI.UpdateProvisionedProduct(&servicecatalog.UpdateProvisionedProductInput{

		ProductId:              aws.String(productID),
		ProvisioningArtifactId: aws.String(artifactID),
		ProvisionedProductName: aws.String(name),
		ProvisioningParameters: provisioningParameters,
		UpdateToken:            aws.String(strconv.FormatInt(time.Now().Unix(), 16)),
	})
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 5)
	stackID, err := scManager.GetStackID(aws.StringValue(ppOut.RecordDetail.RecordId))
	if err != nil {
		return err
	}

	log.Infof("  Updating stack '%s'", stackID)
	scManager.stackManager.AwaitFinalStatus(strings.Split(stackID, "/")[1])
	return nil
}

func (scManager *serviceCatalogManager) UpsertProvisionedProduct(productID string, version string, name string, params map[string]string) error {

	output, err := scManager.scAPI.ListProvisioningArtifacts(&servicecatalog.ListProvisioningArtifactsInput{
		ProductId: aws.String(productID),
	})
	if err != nil {
		return err
	}
	var artifactID string
	for _, artifact := range output.ProvisioningArtifactDetails {
		if aws.BoolValue(artifact.Active) && aws.StringValue(artifact.Name) == version {
			artifactID = aws.StringValue(artifact.Id)
			break
		}
	}

	if artifactID == "" {
		return fmt.Errorf("Unable to find active version '%s' for product '%s'", version, productID)
	}

	provisionedProducts, err := scManager.scAPI.SearchProvisionedProducts(&servicecatalog.SearchProvisionedProductsInput{
		AccessLevelFilter: &servicecatalog.AccessLevelFilter{
			Key:   aws.String(servicecatalog.AccessLevelFilterKeyAccount),
			Value: aws.String("self"),
		},
		Filters: map[string][]*string{
			"SearchQuery": {aws.String(fmt.Sprintf("name:%s", name))},
		},
	})

	for _, provisionedProduct := range provisionedProducts.ProvisionedProducts {
		if name == aws.StringValue(provisionedProduct.Name) {
			return scManager.updateProvisionedProduct(productID, artifactID, name, params)
		}
	}

	return scManager.createProvisionedProduct(productID, artifactID, name, params)
}

func (scManager *serviceCatalogManager) TerminateProvisionedProducts(productID string) error {
	var err error
	err2 := scManager.scAPI.SearchProvisionedProductsPages(&servicecatalog.SearchProvisionedProductsInput{
		AccessLevelFilter: &servicecatalog.AccessLevelFilter{
			Key:   aws.String(servicecatalog.AccessLevelFilterKeyAccount),
			Value: aws.String("self"),
		},
		Filters: map[string][]*string{
			"SearchQuery": {aws.String(fmt.Sprintf("productId:%s", productID))},
		},
	}, func(output *servicecatalog.SearchProvisionedProductsOutput, isLast bool) bool {
		for _, provisionedProduct := range output.ProvisionedProducts {
			if scManager.dryrun {
				log.Infof("  DRYRUN: Skipping termination of provisionedProduct '%s'", aws.StringValue(provisionedProduct.Id))
				continue
			}

			log.Infof("  Deleting provisionedProduct '%s'", aws.StringValue(provisionedProduct.Id))
			stackID := aws.StringValue(provisionedProduct.PhysicalId)

			_, err = scManager.scAPI.TerminateProvisionedProduct(&servicecatalog.TerminateProvisionedProductInput{
				ProvisionedProductId: provisionedProduct.Id,
				TerminateToken:       aws.String(strconv.FormatInt(time.Now().Unix(), 16)),
			})
			if err != nil {
				return false
			}

			time.Sleep(time.Second * 5)

			log.Infof("  Deleting stack '%s'", stackID)
			scManager.stackManager.AwaitFinalStatus(strings.Split(stackID, "/")[1])
		}
		return true
	})
	if err2 != nil {
		return err2
	}
	return err
}

func (scManager *serviceCatalogManager) SetProductVersions(productID string, productVersions map[string]string) error {
	var output *servicecatalog.ListProvisioningArtifactsOutput
	if scManager.dryrun && productID == "" {
		output = &servicecatalog.ListProvisioningArtifactsOutput{
			ProvisioningArtifactDetails: make([]*servicecatalog.ProvisioningArtifactDetail, 0),
		}
	} else {
		input := &servicecatalog.ListProvisioningArtifactsInput{
			ProductId: aws.String(productID),
		}
		var err error
		output, err = scManager.scAPI.ListProvisioningArtifacts(input)
		if err != nil {
			return err
		}
	}

	versionsToAdd := common.MapClone(productVersions)
	for _, artifact := range output.ProvisioningArtifactDetails {
		version := aws.StringValue(artifact.Name)
		active := aws.BoolValue(artifact.Active)
		if _, ok := versionsToAdd[version]; ok {
			// version in SC and in config, make sure it is active
			delete(versionsToAdd, version)

			if !active {
				if scManager.dryrun {
					log.Infof("  DRYRUN: Skipping unarchiving of productVersion '%s'", version)
					continue
				}
				log.Infof("  Unarchiving productVersion '%s'", version)
				_, err := scManager.scAPI.UpdateProvisioningArtifact(&servicecatalog.UpdateProvisioningArtifactInput{
					Active:                 aws.Bool(true),
					ProductId:              aws.String(productID),
					ProvisioningArtifactId: artifact.Id,
				})
				if err != nil {
					return err
				}
			}
		} else {
			// version in SC but not in config, make sure it is inactive

			if active {
				if scManager.dryrun {
					log.Infof("  DRYRUN: Skipping archiving of productVersion '%s'", version)
					continue
				}

				log.Infof("  Archiving productVersion '%s'", version)
				_, err := scManager.scAPI.UpdateProvisioningArtifact(&servicecatalog.UpdateProvisioningArtifactInput{
					Active:                 aws.Bool(false),
					ProductId:              aws.String(productID),
					ProvisioningArtifactId: artifact.Id,
				})
				if err != nil {
					return err
				}
			}
		}

	}

	// add new versions
	for version, templateURL := range versionsToAdd {
		if scManager.dryrun {
			log.Infof("  DRYRUN: Skipping creation of productVersion '%s'", version)
			continue
		}

		log.Infof("  Creating productVersion '%s'", version)
		_, err := scManager.scAPI.CreateProvisioningArtifact(&servicecatalog.CreateProvisioningArtifactInput{
			ProductId: aws.String(productID),
			Parameters: &servicecatalog.ProvisioningArtifactProperties{
				Name: aws.String(version),
				Type: aws.String(servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
				Info: map[string]*string{
					"LoadTemplateFromURL": aws.String(templateURL),
				},
			},
			IdempotencyToken: aws.String(strconv.FormatInt(time.Now().Unix(), 16)),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
