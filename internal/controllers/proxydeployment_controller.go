/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"hash/fnv"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	shulkermciov1alpha1 "github.com/iamblueslime/shulker/api/v1alpha1"
	common "github.com/iamblueslime/shulker/internal/resources"
	resources "github.com/iamblueslime/shulker/internal/resources/proxydeployment"
)

const templateHashLabel = "proxydeployment.shulkermc.io/template-hash"

// ProxyDeploymentReconciler reconciles a ProxyDeployment object
type ProxyDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=shulkermc.io,resources=proxies,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=shulkermc.io,resources=proxydeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=shulkermc.io,resources=proxydeployments/status,verbs=get;update;patch

func (r *ProxyDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling ProxyDeployment")
	proxyDeployment, err := r.getProxyDeployment(ctx, req.NamespacedName)

	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else if k8serrors.IsNotFound(err) {
		// No need to requeue if the resource no longer exists
		return ctrl.Result{}, nil
	}

	cluster := &shulkermciov1alpha1.MinecraftCluster{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: proxyDeployment.Namespace,
		Name:      proxyDeployment.Spec.ClusterRef.Name,
	}, cluster)
	if err != nil {
		logger.Error(err, "Referenced MinecraftCluster does not exists")
		return ctrl.Result{}, err
	}

	resourceBuilder := resources.ProxyDeploymentResourceBuilder{
		Instance: proxyDeployment,
		Scheme:   r.Scheme,
	}
	builders, dirtyBuilders := resourceBuilder.ResourceBuilders()

	err = ReconcileWithResourceBuilders(r.Client, ctx, builders, dirtyBuilders)
	if err != nil {
		return ctrl.Result{}, err
	}

	allProxies, err := r.getAllProxies(ctx, proxyDeployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	if len(allProxies.Items) > int(proxyDeployment.Spec.Replicas) {
		nbToRemove := len(allProxies.Items) - int(proxyDeployment.Spec.Replicas)

		for i := 0; i < nbToRemove; i += 1 {
			err = r.Delete(ctx, &allProxies.Items[i])
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	} else if len(allProxies.Items) < int(proxyDeployment.Spec.Replicas) {
		templateHash := getProxyTemplateHash(&proxyDeployment.Spec.Template)
		matchingProxies, err := r.getMatchingProxies(ctx, proxyDeployment, templateHash)
		if err != nil {
			return ctrl.Result{}, err
		}

		proxiesToCreate := int(proxyDeployment.Spec.Replicas) - len(matchingProxies.Items)
		for i := 0; i < proxiesToCreate; i += 1 {
			proxyId := common.RandomResourceId(6)

			proxy := shulkermciov1alpha1.Proxy{}

			labels := r.getProxyLabelsWithTemplateHash(proxyDeployment, templateHash)
			for k, v := range proxyDeployment.Spec.Template.Labels {
				labels[k] = v
			}

			proxy.Namespace = proxyDeployment.Namespace
			proxy.Name = fmt.Sprintf("%s-%s-%s", proxyDeployment.Name, templateHash, proxyId)
			proxy.Labels = labels
			proxy.Spec = proxyDeployment.Spec.Template.Spec
			proxy.Spec.ClusterRef = proxyDeployment.Spec.ClusterRef
			proxy.Spec.Configuration = shulkermciov1alpha1.ProxyConfigurationSpec{
				ExistingConfigMapName: resourceBuilder.GetConfigMapName(),
			}
			proxy.Spec.PodOverrides.ServiceAccountName = resourceBuilder.GetServiceAccountName()

			if err := controllerutil.SetControllerReference(proxyDeployment, &proxy, r.Scheme); err != nil {
				err = fmt.Errorf("failed setting controller reference for Proxy: %v", err)
				return ctrl.Result{}, err
			}

			err = r.Create(ctx, &proxy)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	allProxies, err = r.getAllProxies(ctx, proxyDeployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	availableReplicas, unavailableReplicas := 0, 0
	for _, proxy := range allProxies.Items {
		for _, condition := range proxy.Status.Conditions {
			if shulkermciov1alpha1.ProxyStatusCondition(condition.Type) == shulkermciov1alpha1.ProxyReadyCondition {
				if condition.Status == metav1.ConditionTrue {
					availableReplicas += 1
				} else {
					unavailableReplicas += 1
				}
			}
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(resourceBuilder.GetPodSelector())
	if err != nil {
		return ctrl.Result{}, err
	}

	proxyDeployment.Status.Replicas = proxyDeployment.Spec.Replicas
	proxyDeployment.Status.AvailableReplicas = int32(availableReplicas)
	proxyDeployment.Status.UnavailableReplicas = int32(unavailableReplicas)
	proxyDeployment.Status.Selector = selector.String()

	if availableReplicas > 0 {
		proxyDeployment.Status.SetCondition(shulkermciov1alpha1.ProxyDeploymentAvailableCondition, metav1.ConditionTrue, "AtLeastOneReady", "One or more proxies are ready")
	} else {
		proxyDeployment.Status.SetCondition(shulkermciov1alpha1.ProxyDeploymentAvailableCondition, metav1.ConditionFalse, "NotReady", "No proxy is ready")
	}

	return ctrl.Result{}, r.Status().Update(ctx, proxyDeployment)
}

func (r *ProxyDeploymentReconciler) getProxyDeployment(ctx context.Context, namespacedName types.NamespacedName) (*shulkermciov1alpha1.ProxyDeployment, error) {
	proxyDeployment := &shulkermciov1alpha1.ProxyDeployment{}
	err := r.Get(ctx, namespacedName, proxyDeployment)
	return proxyDeployment, err
}

func (r *ProxyDeploymentReconciler) getProxyLabels(deployment *shulkermciov1alpha1.ProxyDeployment) map[string]string {
	labels := map[string]string{
		"minecraftcluster.shulkermc.io/name": deployment.Spec.ClusterRef.Name,
		"proxydeployment.shulkermc.io/name":  deployment.Name,
	}
	return labels
}

func (r *ProxyDeploymentReconciler) getProxyLabelsWithTemplateHash(deployment *shulkermciov1alpha1.ProxyDeployment, templateHash string) map[string]string {
	labels := r.getProxyLabels(deployment)
	labels[templateHashLabel] = templateHash
	return labels
}

func (r *ProxyDeploymentReconciler) getAllProxies(ctx context.Context, deployment *shulkermciov1alpha1.ProxyDeployment) (*shulkermciov1alpha1.ProxyList, error) {
	list := shulkermciov1alpha1.ProxyList{}
	err := r.List(ctx, &list, client.InNamespace(deployment.Namespace), client.MatchingLabels(r.getProxyLabels(deployment)))

	return &list, err
}

func (r *ProxyDeploymentReconciler) getMatchingProxies(ctx context.Context, deployment *shulkermciov1alpha1.ProxyDeployment, templateHash string) (*shulkermciov1alpha1.ProxyList, error) {
	list := shulkermciov1alpha1.ProxyList{}
	err := r.List(ctx, &list, client.InNamespace(deployment.Namespace), client.MatchingLabels(r.getProxyLabelsWithTemplateHash(deployment, templateHash)))

	return &list, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProxyDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &shulkermciov1alpha1.ProxyDeployment{}, ".spec.clusterRef.name", func(object client.Object) []string {
		proxyDeployment := object.(*shulkermciov1alpha1.ProxyDeployment)
		return []string{proxyDeployment.Spec.ClusterRef.Name}
	})

	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&shulkermciov1alpha1.ProxyDeployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&shulkermciov1alpha1.Proxy{}).
		Complete(r)
}

func getProxyTemplateHash(template *shulkermciov1alpha1.ProxyTemplate) string {
	hasher := fnv.New32a()
	hashutil.DeepHashObject(hasher, *template)

	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32()))
}
