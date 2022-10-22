package io.shulkermc.proxyagent.adapters.kubernetes.models

import io.fabric8.kubernetes.api.model.DefaultKubernetesResourceList
import io.fabric8.kubernetes.api.model.KubernetesResource
import io.fabric8.kubernetes.api.model.Namespaced
import io.fabric8.kubernetes.client.CustomResource
import io.fabric8.kubernetes.model.annotation.Group
import io.fabric8.kubernetes.model.annotation.Kind
import io.fabric8.kubernetes.model.annotation.Plural
import io.fabric8.kubernetes.model.annotation.Version

private const val GROUP = "shulkermc.io"
private const val VERSION = "v1alpha1"
private const val KIND = "Proxy"
private const val PLURAL = "proxies"

@Group(GROUP)
@Version(VERSION)
@Kind(KIND)
@Plural(PLURAL)
class ShulkerV1alpha1Proxy : CustomResource<ShulkerV1alpha1Proxy.Spec, Void>(), Namespaced {
    class Spec : KubernetesResource
    class List : DefaultKubernetesResourceList<ShulkerV1alpha1Proxy?>()
}
