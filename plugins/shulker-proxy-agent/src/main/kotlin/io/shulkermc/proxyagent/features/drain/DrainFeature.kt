package io.shulkermc.proxyagent.features.drain

import com.velocitypowered.api.event.PostOrder
import com.velocitypowered.api.event.Subscribe
import com.velocitypowered.api.event.connection.PreLoginEvent
import io.shulkermc.proxyagent.ShulkerProxyAgent
import io.shulkermc.proxyagent.adapters.filesystem.FileSystemAdapter
import io.shulkermc.proxyagent.adapters.kubernetes.KubernetesGatewayAdapter
import io.shulkermc.proxyagent.adapters.kubernetes.WatchAction
import io.shulkermc.proxyagent.common.createDisconnectMessage
import net.kyori.adventure.text.format.NamedTextColor
import java.io.IOException
import java.util.concurrent.TimeUnit

class DrainFeature(
        private val agent: ShulkerProxyAgent,
        private val fileSystem: FileSystemAdapter,
        private val kubernetesGateway: KubernetesGatewayAdapter,
        private val ttlSeconds: Long
) {
    companion object {
        const val PROXY_DRAIN_ANNOTATION = "proxy.shulkermc.io/drain"

        val MSG_NOT_ACCEPTING_PLAYERS = createDisconnectMessage(
            "Proxy is not accepting players, try reconnect.",
            NamedTextColor.RED)
    }

    private var acceptingPlayers = true
    private var drained = false

    init {
        this.agent.server.eventManager.register(agent.plugin, this)

        kubernetesGateway.watchProxyEvent { action, proxy ->
            this.agent.logger.fine("Detected modification on Kubernetes Proxy ${proxy.metadata.name}")

            if (action == WatchAction.MODIFIED) {
                val annotations: Map<String, String> = proxy.metadata.annotations
                    ?: return@watchProxyEvent

                if (annotations.containsKey(PROXY_DRAIN_ANNOTATION))
                    if (annotations[PROXY_DRAIN_ANNOTATION] == "true")
                        this.drain()
            }
        }

        this.agent.logger.info("Proxy will be force stopped in ${this.ttlSeconds} seconds")
        this.agent.server.scheduler.buildTask(this.agent.plugin) {
            this.agent.server.shutdown()
        }.delay(this.ttlSeconds, TimeUnit.SECONDS).schedule()
    }

    @Subscribe(order = PostOrder.FIRST)
    fun onPreLogin(event: PreLoginEvent) {
        if (!this.acceptingPlayers)
            event.result = PreLoginEvent.PreLoginComponentResult.denied(MSG_NOT_ACCEPTING_PLAYERS)
    }

    private fun drain() {
        if (this.drained)
            return
        this.drained = true

        try {
            this.fileSystem.createDrainFile()
            this.acceptingPlayers = false
            this.kubernetesGateway.emitNotAcceptingPlayers()
        } catch (e: IOException) {
            throw RuntimeException(e)
        }

        this.agent.logger.info("Proxy is no longer accepting players");

        this.agent.server.scheduler.buildTask(this.agent.plugin) {
            val playerCount = this.agent.server.playerCount

            if (playerCount == 0) {
                this.agent.logger.info("Proxy is empty, stopping")
                this.agent.server.shutdown()
            } else {
                this.agent.logger.info(String.format("There are still %d players connected, waiting", playerCount))
            }
        }.repeat(30L, TimeUnit.SECONDS).schedule()
    }
}
