package io.shulkermc.proxyagent.features.directory

import com.velocitypowered.api.proxy.server.ServerInfo
import io.shulkermc.proxyagent.ShulkerProxyAgent
import io.shulkermc.proxyagent.adapters.kubernetes.KubernetesGatewayAdapter
import io.shulkermc.proxyagent.adapters.kubernetes.WatchAction
import io.shulkermc.proxyagent.adapters.kubernetes.models.ShulkerV1alpha1MinecraftServer
import java.net.InetSocketAddress
import java.util.logging.Level
import kotlin.jvm.optionals.getOrElse

@OptIn(ExperimentalStdlibApi::class)
class DirectoryFeature(
    private val agent: ShulkerProxyAgent,
    kubernetesGateway: KubernetesGatewayAdapter,
): DirectoryFeatureAdapter {
    private val serversByTag = HashMap<String, MutableSet<ServerInfo>>()

    init {
        kubernetesGateway.watchMinecraftServerEvent { action, minecraftServer ->
            agent.logger.fine("Detected modification on Kubernetes MinecraftServer ${minecraftServer.metadata.name}")
            if (action == WatchAction.ADDED || action == WatchAction.MODIFIED)
                this.registerServer(minecraftServer)
            else if (action == WatchAction.DELETED)
                this.unregisterServer(minecraftServer)
        }

        val existingMinecraftServers = kubernetesGateway.listMinecraftServers()
        existingMinecraftServers.items
            .filterNotNull()
            .forEach { minecraftServer -> registerServer(minecraftServer) }
    }

    override fun getServersByTag(tag: String): Set<ServerInfo> {
        return this.serversByTag.getOrElse(tag) { HashSet() }
    }

    private fun registerServer(minecraftServer: ShulkerV1alpha1MinecraftServer) {
        val alreadyKnown = this.agent.server.getServer(minecraftServer.metadata.name).isPresent

        if (alreadyKnown || minecraftServer.status == null)
            return

        val readyCondition = minecraftServer.status.getConditionByType("Ready")
        val isReady = readyCondition.map { condition ->
            condition.status == "True"
        }.getOrElse { false }

        if (isReady) {
            this.agent.logger.info("Adding MinecraftServer ${minecraftServer.metadata.name} to directory")
            val serverInfo = ServerInfo(
                minecraftServer.metadata.name,
                InetSocketAddress(minecraftServer.status.serverIP, 25565)
            )

            this.agent.server.registerServer(serverInfo)

            if (minecraftServer.spec.tags != null)
                for (tag in minecraftServer.spec.tags!!)
                    serversByTag.getOrPut(tag) { HashSet() }.add(serverInfo)
        }
    }

    private fun unregisterServer(minecraftServer: ShulkerV1alpha1MinecraftServer) {
        this.agent.logger.info("Removing MinecraftServer ${minecraftServer.metadata.name} from directory")
        this.agent.server.getServer(minecraftServer.metadata.name).ifPresent { server ->
            this.agent.server.unregisterServer(server.serverInfo)
        }
    }
}
