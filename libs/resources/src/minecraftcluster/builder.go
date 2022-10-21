/*
Copyright (c) Jérémy Levilain
SPDX-License-Identifier: GPL-3.0-or-later
*/

package resources

import (
	"fmt"

	shulkermciov1alpha1 "github.com/iamblueslime/shulker/libs/crds/v1alpha1"
	common "github.com/iamblueslime/shulker/libs/resources/src"
	"k8s.io/apimachinery/pkg/runtime"
)

type MinecraftClusterResourceBuilder struct {
	Instance *shulkermciov1alpha1.MinecraftCluster
	Scheme   *runtime.Scheme
}

func (b *MinecraftClusterResourceBuilder) ResourceBuilders() ([]common.ResourceBuilder, []common.ResourceBuilder) {
	builders := []common.ResourceBuilder{
		b.MinecraftClusterProxyServiceAccount(),
		b.MinecraftClusterProxyWatchRole(),
		b.MinecraftClusterProxyWatchRoleBinding(),
	}
	dirtyBuilders := []common.ResourceBuilder{}

	return builders, dirtyBuilders
}

func (b *MinecraftClusterResourceBuilder) getProxyServiceAccountName() string {
	return fmt.Sprintf("%s-proxy", b.Instance.Name)
}

func (b *MinecraftClusterResourceBuilder) getProxyWatchRoleName() string {
	return fmt.Sprintf("%s-proxy-watch", b.Instance.Name)
}

func (b *MinecraftClusterResourceBuilder) getProxyWatchRoleBindingName() string {
	return fmt.Sprintf("%s-proxy-watch", b.Instance.Name)
}

func (b *MinecraftClusterResourceBuilder) getLabels() map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/name":             b.Instance.Name,
		"app.kubernetes.io/component":        "cluster",
		"minecraftcluster.shulkermc.io/name": b.Instance.Name,
	}

	return labels
}
