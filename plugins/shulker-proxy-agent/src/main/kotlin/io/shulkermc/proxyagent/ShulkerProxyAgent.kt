package io.shulkermc.proxyagent

import io.shulkermc.proxyagent.adapters.filesystem.FileSystemAdapterImpl
import io.shulkermc.proxyagent.adapters.kubernetes.KubernetesGatewayAdapterImpl
import io.shulkermc.proxyagent.features.drain.DrainFeature
import net.md_5.bungee.api.plugin.Plugin

class ShulkerProxyAgent : Plugin() {
    override fun onEnable() {
        val config = parse()
        this.logger.info(String.format("Identified Shulker proxy: %s/%s", config.proxyNamespace, config.proxyName))

        val fileSystem = FileSystemAdapterImpl()
        val kubernetesGateway = KubernetesGatewayAdapterImpl(config.proxyNamespace, config.proxyName)

        DrainFeature(this, fileSystem, kubernetesGateway)

        kubernetesGateway.emitAgentReady()
    }
}
