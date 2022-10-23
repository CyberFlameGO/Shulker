package io.shulkermc.proxyagent.features.directory

import com.velocitypowered.api.proxy.server.ServerInfo

interface DirectoryFeatureAdapter {
    fun getServersByTag(tag: String): Set<ServerInfo>
}
