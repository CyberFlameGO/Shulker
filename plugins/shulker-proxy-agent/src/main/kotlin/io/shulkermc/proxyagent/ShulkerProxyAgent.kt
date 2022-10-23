package io.shulkermc.proxyagent

import com.velocitypowered.api.event.proxy.ProxyInitializeEvent
import com.velocitypowered.api.event.proxy.ProxyShutdownEvent
import com.velocitypowered.api.proxy.ProxyServer
import io.shulkermc.proxyagent.adapters.filesystem.FileSystemAdapterImpl
import io.shulkermc.proxyagent.adapters.kubernetes.KubernetesGatewayAdapter
import io.shulkermc.proxyagent.adapters.kubernetes.KubernetesGatewayAdapterImpl
import io.shulkermc.proxyagent.features.directory.DirectoryFeature
import io.shulkermc.proxyagent.features.drain.DrainFeature
import io.shulkermc.proxyagent.features.limbo.LimboFeature
import java.lang.Exception
import java.util.logging.Logger

class ShulkerProxyAgent(var plugin: BootstrapPlugin, var server: ProxyServer, var logger: Logger) {
    private var kubernetesGateway: KubernetesGatewayAdapter? = null

    fun onProxyInitialization(@Suppress("UNUSED_PARAMETER") event: ProxyInitializeEvent) {
        try {
            val config = parse()

            this.logger.info("Identified Shulker proxy: ${config.proxyNamespace}/${config.proxyName}")

            val fileSystem = FileSystemAdapterImpl()
            this.kubernetesGateway = KubernetesGatewayAdapterImpl(config.proxyNamespace, config.proxyName)

            DrainFeature(this, fileSystem, kubernetesGateway!!, config.ttlSeconds)
            val directoryFeature = DirectoryFeature(this, kubernetesGateway!!)
            LimboFeature(this, directoryFeature)

            kubernetesGateway!!.emitAgentReady()
        } catch (e: Exception) {
            this.logger.severe("Failed to parse configuration")
            e.printStackTrace()
            server.shutdown()
        }
    }

    fun onProxyShutdown(@Suppress("UNUSED_PARAMETER") event: ProxyShutdownEvent) {
        if (this.kubernetesGateway != null)
            this.kubernetesGateway!!.destroy()
    }
}
