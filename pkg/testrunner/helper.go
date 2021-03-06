package testrunner

import (
	"context"
	"fmt"
	"sort"

	"github.com/Masterminds/semver"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getLatestK8sVersion(gardenerKubeconfigPath, cloudprofile, cloudprovider string) (string, error) {
	ctx := context.Background()
	defer ctx.Done()
	k8sGardenClient, err := kubernetes.NewClientFromFile(gardenerKubeconfigPath, nil, client.Options{
		Scheme: kubernetes.GardenScheme,
	})
	if err != nil {
		return "", err
	}

	profile := &gardenv1beta1.CloudProfile{}
	err = k8sGardenClient.Client().Get(ctx, types.NamespacedName{Name: cloudprofile}, profile)
	if err != nil {
		return "", err
	}

	rawVersions, err := getCloudproviderVersions(profile, cloudprovider)
	if err != nil {
		return "", err
	}

	if len(rawVersions) == 0 {
		return "", fmt.Errorf("No kubernetes versions found for cloudprofle %s", cloudprofile)
	}

	versions := make([]*semver.Version, len(rawVersions))
	for i, rawVersion := range rawVersions {
		v, err := semver.NewVersion(rawVersion)
		if err == nil {
			versions[i] = v
		}
	}
	sort.Sort(semver.Collection(versions))

	return versions[len(versions)-1].String(), nil
}

func getCloudproviderVersions(profile *gardenv1beta1.CloudProfile, cloudprovider string) ([]string, error) {

	switch gardenv1beta1.CloudProvider(cloudprovider) {
	case gardenv1beta1.CloudProviderAWS:
		return profile.Spec.AWS.Constraints.Kubernetes.Versions, nil
	case gardenv1beta1.CloudProviderGCP:
		return profile.Spec.GCP.Constraints.Kubernetes.Versions, nil
	case gardenv1beta1.CloudProviderAzure:
		return profile.Spec.Azure.Constraints.Kubernetes.Versions, nil
	case gardenv1beta1.CloudProviderOpenStack:
		return profile.Spec.OpenStack.Constraints.Kubernetes.Versions, nil
	case gardenv1beta1.CloudProviderAlicloud:
		return profile.Spec.Alicloud.Constraints.Kubernetes.Versions, nil
	default:
		return nil, fmt.Errorf("Unsupported cloudprovider %s", cloudprovider)
	}
}
