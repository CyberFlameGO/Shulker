package io.shulkermc.proxyagent.adapters.kubernetes

import io.shulkermc.proxyagent.adapters.kubernetes.models.ShulkerV1alpha1Proxy

enum class WatchAction {
    ADDED, MODIFIED, DELETED
}

interface KubernetesGatewayAdapter {
    fun emitAgentReady()
    fun emitNotAcceptingPlayers()

    fun watchProxyEvent(callback: (action: WatchAction, proxy: ShulkerV1alpha1Proxy) -> Unit)
}
