package resources

import (
	shulkermciov1alpha1 "github.com/iamblueslime/shulker/api/v1alpha1"
	common "github.com/iamblueslime/shulker/internal/resources"
	"k8s.io/apimachinery/pkg/runtime"
)

type MinecraftClusterResourceBuilder struct {
	Instance *shulkermciov1alpha1.MinecraftCluster
	Scheme   *runtime.Scheme
}

func (b *MinecraftClusterResourceBuilder) ResourceBuilders() ([]common.ResourceBuilder, []common.ResourceBuilder) {
	builders := []common.ResourceBuilder{}
	dirtyBuilders := []common.ResourceBuilder{}

	return builders, dirtyBuilders
}
