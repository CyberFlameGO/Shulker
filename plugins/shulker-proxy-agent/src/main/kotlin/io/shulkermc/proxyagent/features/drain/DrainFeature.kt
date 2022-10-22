package io.shulkermc.proxyagent.features.drain

import io.shulkermc.proxyagent.ShulkerProxyAgent
import io.shulkermc.proxyagent.adapters.filesystem.FileSystemAdapter
import io.shulkermc.proxyagent.adapters.kubernetes.KubernetesGatewayAdapter
import io.shulkermc.proxyagent.adapters.kubernetes.WatchAction
import io.shulkermc.proxyagent.common.createDisconnectMessage
import net.md_5.bungee.api.ChatColor
import net.md_5.bungee.api.ProxyServer
import net.md_5.bungee.api.event.ServerConnectEvent
import net.md_5.bungee.api.plugin.Listener
import net.md_5.bungee.event.EventHandler
import net.md_5.bungee.event.EventPriority
import java.io.IOException
import java.util.concurrent.TimeUnit
import java.util.logging.Logger

private var PROXY_DRAIN_ANNOTATION = "proxy.shulkermc.io/drain"

private var DRAIN_DELAY_SECONDS = 10L
private var STOP_IF_EMPTY_DELAY_SECONDS = 30L
private var STOP_FORCE_DELAY_SECONDS = 60L * 60L * 3L

private var MSG_NOT_ACCEPTING_PLAYERS = createDisconnectMessage(
    "Proxy is not accepting players, try reconnect.",
    ChatColor.RED)

class DrainFeature(
        private val plugin: ShulkerProxyAgent,
        private val fileSystem: FileSystemAdapter,
        private val kubernetesGateway: KubernetesGatewayAdapter
): Listener {
    private val logger: Logger = plugin.logger
    private val proxyServer: ProxyServer = plugin.proxy

    private var acceptingPlayers = true
    private var drained = false

    init {
        this.proxyServer.pluginManager.registerListener(plugin, this)

        this.proxyServer.scheduler.runAsync(plugin, Runnable {
            kubernetesGateway.watchProxyEvent { action, proxy ->
                if (action == WatchAction.MODIFIED) {
                    this.logger.info("Detected modification on Kubernetes Proxy")

                    val annotations: Map<String, String> = proxy.metadata.annotations
                        ?: return@watchProxyEvent

                    if (annotations.containsKey(PROXY_DRAIN_ANNOTATION)) {
                        val drainValue = annotations[PROXY_DRAIN_ANNOTATION]

                        this.logger.info(String.format("Found drain annotation: %s=%b", PROXY_DRAIN_ANNOTATION, drainValue))

                        if (drainValue == "true") {
                            this.logger.info("Invoking drain callbacks")
                            this.drain()
                        }
                    }
                }
            }
        })
    }

    @EventHandler(priority = EventPriority.HIGHEST)
    fun onServerConnect(event: ServerConnectEvent) {
        if (!this.acceptingPlayers) {
            event.player.disconnect(*MSG_NOT_ACCEPTING_PLAYERS)
            event.isCancelled = true
        }
    }

    private fun drain() {
        if (this.drained)
            return;
        this.drained = true;

        this.logger.info("Proxy will be force stopped in $STOP_FORCE_DELAY_SECONDS seconds")

        this.proxyServer.scheduler.schedule(this.plugin, {
            this.logger.info("Proxy is no longer accepting players");

            try {
                this.fileSystem.createDrainFile()
                this.acceptingPlayers = false;
                this.kubernetesGateway.emitNotAcceptingPlayers()
            } catch (e: IOException) {
                throw RuntimeException(e)
            }
        }, DRAIN_DELAY_SECONDS, TimeUnit.SECONDS)

        this.proxyServer.scheduler.schedule(this.plugin, {
            val playersLeft = this.proxyServer.players.size

            if (playersLeft == 0) {
                this.logger.info("Proxy is empty, stopping")
                this.proxyServer.stop()
            } else {
                this.logger.info(String.format("There are still %d players connected, waiting", playersLeft))
            }
        }, STOP_IF_EMPTY_DELAY_SECONDS, STOP_IF_EMPTY_DELAY_SECONDS, TimeUnit.SECONDS)

        this.proxyServer.scheduler.schedule(this.plugin, {
            this.proxyServer.stop()
        }, STOP_FORCE_DELAY_SECONDS, TimeUnit.SECONDS)
    }
}
