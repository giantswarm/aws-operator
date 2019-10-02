package helmclient

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/errors/tenant"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/helm/cmd/helm/installer"
)

// EnsureTillerInstalled installs Tiller by creating its deployment and waiting
// for it to start. A service account and cluster role binding are also created.
// As a first step, it checks if Tiller is already ready, in which case it
// returns early.
func (c *Client) EnsureTillerInstalled(ctx context.Context) error {
	return c.EnsureTillerInstalledWithValues(ctx, []string{})
}

// EnsureTillerInstalledWithValues installs Tiller by creating its deployment
// and waiting for it to start. A service account and cluster role binding are
// also created. As a first step, it checks if Tiller is already ready, in
// which case it returns early. Values can be provided to pass through to Tiller
// and overwrite its deployment.
func (c *Client) EnsureTillerInstalledWithValues(ctx context.Context, values []string) error {
	// Check if Tiller is already present and return early if so.
	{
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding if tiller is installed in namespace %#q", c.tillerNamespace))

		t, err := c.newTunnel()
		defer c.closeTunnel(ctx, t)
		if err != nil {
			// fall through, we may need to create or upgrade Tiller.
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found that tiller is not installed in namespace %#q", c.tillerNamespace))
		} else {
			err = c.newHelmClientFromTunnel(t).PingTiller()
			if err == nil {
				c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found that tiller is installed in namespace %#q", c.tillerNamespace))
				return nil
			}
		}
	}

	// Create the service account for tiller so it can pull images and do its do.
	{
		name := tillerPodName
		namespace := c.tillerNamespace

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating serviceaccount %#q in namespace %#q", name, namespace))

		serviceAccont := &corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}

		_, err := c.k8sClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccont)
		if errors.IsAlreadyExists(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("serviceaccount %#q in namespace %#q already exists", name, namespace))
			// fall through
		} else if tenant.IsAPINotAvailable(err) {
			return microerror.Maskf(tenant.APINotAvailableError, err.Error())
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created serviceaccount %#q in namespace %#q", name, namespace))
		}
	}

	// Create the cluster role binding for tiller so it is allowed to do its job.
	{
		serviceAccountName := tillerPodName
		serviceAccountNamespace := c.tillerNamespace

		name := fmt.Sprintf("%s-%s", roleBindingNamePrefix, serviceAccountNamespace)

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating clusterrolebinding %#q", name))

		i := &rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      serviceAccountName,
					Namespace: serviceAccountNamespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			},
		}

		_, err := c.k8sClient.RbacV1().ClusterRoleBindings().Create(i)
		if errors.IsAlreadyExists(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("clusterrolebinding %#q already exists", name))
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created clusterrolebinding %#q", name))
		}
	}

	// Create the network policy for tiller so it is allowed to do its job in case all traffic is blocked.
	{
		networkPolicyName := tillerPodName
		networkPolicyNamespace := c.tillerNamespace
		protocolTCP := corev1.ProtocolTCP
		tillerPort := intstr.IntOrString{
			IntVal: 44134,
		}
		tillerHTTPPort := intstr.IntOrString{
			IntVal: 44135,
		}

		name := fmt.Sprintf("%s/%s", networkPolicyNamespace, networkPolicyName)

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating networkpolicy %#q", name))

		np := &networkingv1.NetworkPolicy{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "NetworkPolicy",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      networkPolicyName,
				Namespace: networkPolicyNamespace,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":  "helm",
						"name": "tiller",
					},
				},
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{
						Ports: []networkingv1.NetworkPolicyPort{
							{
								Protocol: &protocolTCP,
								Port:     &tillerPort,
							},
							{
								Protocol: &protocolTCP,
								Port:     &tillerHTTPPort,
							},
						},
					},
				},
				Egress: []networkingv1.NetworkPolicyEgressRule{
					{},
				},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
					networkingv1.PolicyTypeEgress,
				},
			},
		}

		_, err := c.k8sClient.NetworkingV1().NetworkPolicies(networkPolicyNamespace).Create(np)
		if errors.IsAlreadyExists(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("networkpolicy %#q already exists", name))
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created networkpolicy %#q", name))
		}
	}

	var err error
	var installTiller bool
	var pod *corev1.Pod
	var upgradeTiller bool

	{
		o := func() error {
			pod, err = getPod(c.k8sClient, c.tillerNamespace)
			if IsNotFound(err) {
				// Fall through as we need to install Tiller.
				installTiller = true
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		b := backoff.NewConstant(c.ensureTillerInstalledMaxWait, 5*time.Second)
		n := backoff.NewNotifier(c.logger, context.Background())

		err = backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if !installTiller && pod != nil {
		err = validateTillerVersion(pod, c.tillerImage)
		if IsTillerInvalidVersion(err) {
			upgradeTiller = true
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	i := &installer.Options{
		AutoMountServiceAccountToken: true,
		ForceUpgrade:                 true,
		ImageSpec:                    c.tillerImage,
		MaxHistory:                   defaultMaxHistory,
		Namespace:                    c.tillerNamespace,
		ServiceAccount:               tillerPodName,
		Values:                       values,
	}

	// Install the tiller deployment in the tenant cluster.
	if installTiller && !upgradeTiller {
		err = c.installTiller(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if !installTiller && upgradeTiller {
		err = c.upgradeTiller(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if installTiller && upgradeTiller {
		return microerror.Maskf(executionFailedError, "invalid state cannot both install and upgrade tiller")
	}

	// Wait for tiller to be up and running. When verifying to be able to ping
	// tiller we make sure 3 consecutive pings succeed before assuming everything
	// is fine.
	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting for tiller to be up")

		var i int

		o := func() error {
			t, err := c.newTunnel()
			if !installTiller && IsTillerNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}
			defer c.closeTunnel(ctx, t)

			err = c.newHelmClientFromTunnel(t).PingTiller()
			if err != nil {
				i = 0
				return microerror.Mask(err)
			}

			i++
			if i < 3 {
				return microerror.Maskf(tillerNotFoundError, "failed to ping tiller 3 consecutive times")
			}

			return nil
		}
		b := backoff.NewExponential(c.ensureTillerInstalledMaxWait, 5*time.Second)
		n := backoff.NewNotifier(c.logger, ctx)

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "waited for tiller to be up")
	}

	return nil
}

func (c *Client) installTiller(ctx context.Context, installerOptions *installer.Options) error {
	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating tiller in namespace %#q", c.tillerNamespace))

	o := func() error {
		err := installer.Install(c.k8sClient, installerOptions)
		if errors.IsAlreadyExists(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("tiller in namespace %#q already exists", c.tillerNamespace))
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created tiller in namespace %#q", c.tillerNamespace))
		}

		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 5*time.Second)
	n := backoff.NewNotifier(c.logger, context.Background())

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) upgradeTiller(ctx context.Context, installerOptions *installer.Options) error {
	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("upgrading tiller in namespace %#q", c.tillerNamespace))

	o := func() error {
		err := installer.Upgrade(c.k8sClient, installerOptions)
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("upgraded tiller in namespace %#q", c.tillerNamespace))

		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 5*time.Second)
	n := backoff.NewNotifier(c.logger, context.Background())

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
