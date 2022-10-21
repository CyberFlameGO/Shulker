package io.shulkermc.proxyagent;

import net.md_5.bungee.api.ProxyServer;
import net.md_5.bungee.api.plugin.Plugin;

public class ShulkerProxyAgent extends Plugin {
    private final ProxyServer proxyServer;

    public ShulkerProxyAgent() {
        this.proxyServer = ProxyServer.getInstance();
    }

    @Override
    public void onEnable() {
    }
}