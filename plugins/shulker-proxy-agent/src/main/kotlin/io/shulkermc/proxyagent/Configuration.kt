package io.shulkermc.proxyagent

data class Configuration(
        val proxyNamespace: String,
        val proxyName: String,
        val ttlSeconds: Int
)

fun parse(): Configuration {
    val proxyNamespace = System.getenv("SHULKER_PROXY_NAMESPACE")
            ?: throw RuntimeException("No SHULKER_PROXY_NAMESPACE found in environment")

    val proxyName = System.getenv("SHULKER_PROXY_NAME")
            ?: throw RuntimeException("No SHULKER_PROXY_NAME found in environment")

    val ttlSecondsStr = System.getenv("SHULKER_PROXY_TTL_SECONDS")
            ?: throw RuntimeException("No SHULKER_PROXY_TTL_SECONDS found in environment")
    val ttlSeconds = ttlSecondsStr.toInt()

    return Configuration(
            proxyNamespace,
            proxyName,
            ttlSeconds
    )
}
