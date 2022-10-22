package io.shulkermc.proxyagent

import com.velocitypowered.api.event.proxy.ProxyInitializeEvent
import com.velocitypowered.api.proxy.ProxyServer
import io.shulkermc.proxyagent.adapters.filesystem.FileSystemAdapterImpl
import io.shulkermc.proxyagent.adapters.kubernetes.KubernetesGatewayAdapterImpl
import io.shulkermc.proxyagent.features.drain.DrainFeature
import java.lang.Exception
import java.util.logging.Logger

class ShulkerProxyAgent(var plugin: BootstrapPlugin, var server: ProxyServer, var logger: Logger) {
    fun onProxyInitialization(event: ProxyInitializeEvent) {
        try {
            val config = parse()

            this.logger.info("Identified Shulker proxy: ${config.proxyNamespace}/${config.proxyName}")

            val fileSystem = FileSystemAdapterImpl()
            val kubernetesGateway = KubernetesGatewayAdapterImpl(config.proxyNamespace, config.proxyName)

            DrainFeature(this, fileSystem, kubernetesGateway, config.ttlSeconds)

            kubernetesGateway.emitAgentReady()
        } catch (e: Exception) {
            this.logger.severe("Failed to parse configuration")
            e.printStackTrace()
            server.shutdown()
        }
    }
}
