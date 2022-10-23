package io.shulkermc.proxyagent.features.limbo

import com.velocitypowered.api.event.PostOrder
import com.velocitypowered.api.event.Subscribe
import com.velocitypowered.api.event.player.ServerPreConnectEvent
import io.shulkermc.proxyagent.ShulkerProxyAgent
import io.shulkermc.proxyagent.common.createDisconnectMessage
import io.shulkermc.proxyagent.features.directory.DirectoryFeatureAdapter
import net.kyori.adventure.text.format.NamedTextColor

class LimboFeature(
    private val agent: ShulkerProxyAgent,
    private val directoryFeature: DirectoryFeatureAdapter
) {
    companion object {
        const val LIMBO_TAG = "limbo"

        val MSG_NO_LIMBO_FOUND = createDisconnectMessage(
            "No limbo server found, please check your cluster configuration.",
            NamedTextColor.RED)
    }
    init {
        this.agent.server.eventManager.register(this.agent.plugin, this)
    }

    @Subscribe(order = PostOrder.LAST)
    fun onServerPreConnect(event: ServerPreConnectEvent) {
        event.result = ServerPreConnectEvent.ServerResult.allowed(this.agent.server.allServers.elementAt(0))
        return

        if (event.originalServer.serverInfo.name == LIMBO_TAG) {
            val limboServerInfos = this.directoryFeature.getServersByTag(LIMBO_TAG).iterator()

            if (limboServerInfos.hasNext()) {
                val firstLimboInfo = limboServerInfos.next()
                val firstLimboRegistered = this.agent.server.getServer(firstLimboInfo.name).get()
                event.result = ServerPreConnectEvent.ServerResult.allowed(firstLimboRegistered)
            } else {
                event.player.disconnect(MSG_NO_LIMBO_FOUND)
            }
        }
    }
}
