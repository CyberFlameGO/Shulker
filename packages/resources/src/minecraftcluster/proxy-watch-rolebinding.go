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
)

type MinecraftClusterProxyWatchRoleBindingBuilder struct {
	*MinecraftClusterResourceBuilder
}

func (b *MinecraftClusterResourceBuilder) MinecraftClusterProxyWatchRoleBinding() *MinecraftClusterProxyWatchRoleBindingBuilder {
	return &MinecraftClusterProxyWatchRoleBindingBuilder{b}
}

func (b *MinecraftClusterProxyWatchRoleBindingBuilder) Build() (client.Object, error) {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.getProxyWatchRoleBindingName(),
			Namespace: b.Instance.Namespace,
			Labels:    b.getLabels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     b.getProxyWatchRoleName(),
		},
	}, nil
}

func (b *MinecraftClusterProxyWatchRoleBindingBuilder) Update(object client.Object) error {
	roleBinding := object.(*rbacv1.RoleBinding)

	roleBinding.Subjects = []rbacv1.Subject{{
		Kind:      "ServiceAccount",
		Name:      b.getProxyServiceAccountName(),
		Namespace: b.Instance.Namespace,
	}}

	if err := controllerutil.SetControllerReference(b.Instance, roleBinding, b.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference for RoleBinding: %v", err)
	}

	return nil
}

func (b *MinecraftClusterProxyWatchRoleBindingBuilder) CanBeUpdated() bool {
	return true
}
