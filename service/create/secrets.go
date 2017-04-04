package create

import (
	"fmt"
	"time"

	"github.com/giantswarm/certificatetpr"
	microerror "github.com/giantswarm/microkit/error"
	"k8s.io/client-go/pkg/api"
	apierrs "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/watch"
)

const secretsTimeOut time.Duration = 90 * time.Second

func (s *Service) getCertsFromSecrets(clusterID string) (certificatetpr.AssetsBundle, error) {
	assetsBundle := make(certificatetpr.AssetsBundle)

	var err error
	for _, componentName := range certificatetpr.ClusterComponents {
		assetsBundle, err = s.getComponentSecret(componentName.String(), clusterID, assetsBundle)

		if err != nil {
			return nil, microerror.MaskAny(err)
		}
	}

	return assetsBundle, nil
}

// gets the secret for a single cluster component, and saves in the shared assets bundle
func (s *Service) getComponentSecret(componentName, clusterID string, bundle certificatetpr.AssetsBundle) (certificatetpr.AssetsBundle, error) {
	watcher, err := s.K8sClient.Core().Secrets(api.NamespaceDefault).Watch(v1.ListOptions{
		// select only secrets that pertain to the component AND have a matching clusterID
		LabelSelector: fmt.Sprintf("%s=%s, %s=%s", certificatetpr.ComponentLabel, componentName, certificatetpr.ClusterIDLabel, clusterID),
	})
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	defer watcher.Stop()
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil, microerror.MaskAnyf(secretsRetrievalFailedError, "secrets channel was already closed")
			}

			switch event.Type {
			case watch.Added:
				secret := event.Object.(*v1.Secret)

				component := certificatetpr.ClusterComponent(secret.Labels[certificatetpr.ComponentLabel])

				if !certificatetpr.ValidComponent(component, certificatetpr.ClusterComponents) {
					return nil, microerror.MaskAnyf(secretsRetrievalFailedError, "unknown clusterComponent %s", component)
				}

				for _, assetType := range certificatetpr.TLSAssetTypes {
					asset, ok := secret.Data[assetType.String()]
					if !ok {
						return nil, microerror.MaskAnyf(secretsRetrievalFailedError, "malformed secret was missing %v asset", assetType)
					}

					bundle[certificatetpr.AssetsBundleKey{component, assetType}] = asset
				}

				return bundle, nil
			case watch.Deleted:
				// noop; ignore deleted events
			case watch.Error:
				return nil, microerror.MaskAnyf(secretsRetrievalFailedError, "there was an error in the watcher: %v", apierrs.FromObject(event.Object))
			}

		case <-time.After(secretsTimeOut):
			return nil, microerror.MaskAnyf(secretsRetrievalFailedError, "timed out waiting for secrets")
		}
	}
}
