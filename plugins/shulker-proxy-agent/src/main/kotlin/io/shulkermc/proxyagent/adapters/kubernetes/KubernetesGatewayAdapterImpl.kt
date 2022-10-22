package io.shulkermc.proxyagent.adapters.kubernetes

import io.fabric8.kubernetes.api.model.EventBuilder
import io.fabric8.kubernetes.api.model.ObjectReference
import io.fabric8.kubernetes.api.model.ObjectReferenceBuilder
import io.fabric8.kubernetes.client.KubernetesClient
import io.fabric8.kubernetes.client.KubernetesClientBuilder
import io.fabric8.kubernetes.client.informers.ResourceEventHandler
import io.fabric8.kubernetes.client.okhttp.OkHttpClientFactory
import io.shulkermc.proxyagent.adapters.kubernetes.models.ShulkerV1alpha1Proxy
import java.time.OffsetDateTime

class KubernetesGatewayAdapterImpl(proxyNamespace: String, proxyName: String) : KubernetesGatewayAdapter {
    var kubernetesClient: KubernetesClient
    var proxyReference: ObjectReference

    init {
        this.kubernetesClient = KubernetesClientBuilder()
                .withHttpClientFactory(OkHttpClientFactory())
                .build()

        this.proxyReference = ObjectReferenceBuilder()
                .withApiVersion("shulkermc.io/v1alpha1")
                .withKind("Proxy")
                .withNamespace(proxyNamespace)
                .withName(proxyName)
                .build()
    }

    override fun emitAgentReady() {
        val event = this.createEventBuilder()
                .withType("Normal")
                .withReason("AgentReady")
                .withMessage("Agent is initialized and ready")
                .build()

        this.kubernetesClient.v1().events()
                .resource(event)
                .create()
    }

    override fun emitNotAcceptingPlayers() {
        val event = this.createEventBuilder()
                .withType("Normal")
                .withReason("NotAcceptingPlayers")
                .withMessage("Proxy is no longer accepting players")
                .build()

        this.kubernetesClient.v1().events()
                .resource(event)
                .create()
    }

    override fun watchProxyEvent(callback: (action: WatchAction, proxy: ShulkerV1alpha1Proxy) -> Unit) {
        val proxyApi = this.kubernetesClient.resources(ShulkerV1alpha1Proxy::class.java, ShulkerV1alpha1Proxy.List::class.java);
        val eventHandler = object : ResourceEventHandler<ShulkerV1alpha1Proxy> {
            override fun onAdd(proxy: ShulkerV1alpha1Proxy) {
                callback(WatchAction.ADDED, proxy);
            }

            override fun onUpdate(oldProxy: ShulkerV1alpha1Proxy, newProxy: ShulkerV1alpha1Proxy) {
                callback(WatchAction.MODIFIED, newProxy);
            }

            override fun onDelete(proxy: ShulkerV1alpha1Proxy, deletedFinalStateUnknown: Boolean) {
                callback(WatchAction.DELETED, proxy);
            }
        };

        val proxyInformer = proxyApi
                .inNamespace(this.proxyReference.namespace)
                .withName(this.proxyReference.name)
                .inform(eventHandler, 30L * 1000);

        proxyInformer.start();
    }

    private fun createEventBuilder(): EventBuilder {
        val timestamp = OffsetDateTime.now().toString();

        return EventBuilder()
                .withNewMetadata()
                    .withNamespace(this.proxyReference.namespace)
                    .withGenerateName(this.proxyReference.name)
                .endMetadata()
                .withNewSource()
                    .withComponent("shulker-proxy-agent")
                .endSource()
                .withInvolvedObject(this.proxyReference)
                .withFirstTimestamp(timestamp)
                .withLastTimestamp(timestamp)
                .withCount(1)
    }
}
