/*
Copyright (c) Jérémy Levilain
SPDX-License-Identifier: GPL-3.0-or-later
*/

package resources

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	shulkermciov1alpha1 "github.com/iamblueslime/shulker/packages/crds/v1alpha1"
)

type MinecraftClusterProxyWatchRoleBuilder struct {
	*MinecraftClusterResourceBuilder
}

func (b *MinecraftClusterResourceBuilder) MinecraftClusterProxyWatchRole() *MinecraftClusterProxyWatchRoleBuilder {
	return &MinecraftClusterProxyWatchRoleBuilder{b}
}

func (b *MinecraftClusterProxyWatchRoleBuilder) Build() (client.Object, error) {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.getProxyWatchRoleName(),
			Namespace: b.Instance.Namespace,
			Labels:    b.getLabels(),
		},
	}, nil
}

func (b *MinecraftClusterProxyWatchRoleBuilder) Update(object client.Object) error {
	role := object.(*rbacv1.Role)

	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{shulkermciov1alpha1.GroupVersion.Group},
			Resources: []string{"proxies"},
			Verbs:     []string{"watch"},
		},
	}

	if err := controllerutil.SetControllerReference(b.Instance, role, b.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference for Role: %v", err)
	}

	return nil
}

func (b *MinecraftClusterProxyWatchRoleBuilder) CanBeUpdated() bool {
	return true
}
