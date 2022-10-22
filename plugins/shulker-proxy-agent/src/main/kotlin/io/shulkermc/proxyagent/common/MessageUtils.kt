package io.shulkermc.proxyagent.common

import net.md_5.bungee.api.ChatColor
import net.md_5.bungee.api.chat.BaseComponent
import net.md_5.bungee.api.chat.ComponentBuilder

fun createDisconnectMessage(message: String, color: ChatColor): Array<BaseComponent> =
        ComponentBuilder()
                .append("◆ Shulker ◆\n")
                .color(ChatColor.LIGHT_PURPLE)
                .bold(true)
                .append(message)
                .color(color)
                .bold(false)
                .create()
