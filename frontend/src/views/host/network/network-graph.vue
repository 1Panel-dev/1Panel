<template>
    <div class="network-canvas">
        <el-empty v-if="!groups.length" />
        <div v-else class="network-map">
            <div class="map-headings">
                <span>{{ $t('networkManage.interface') }} / IP</span>
                <span>{{ $t('networkManage.routes') }}</span>
            </div>
            <section v-for="group in groups" :key="group.name" class="network-group">
                <div
                    class="interface-node"
                    :class="{ selectable: group.iface }"
                    :role="group.iface ? 'button' : undefined"
                    :tabindex="group.iface ? 0 : undefined"
                    @click="group.iface && $emit('select-interface', group.name)"
                    @keydown.enter.self.prevent="group.iface && $emit('select-interface', group.name)"
                    @keydown.space.self.prevent="group.iface && $emit('select-interface', group.name)"
                >
                    <div class="interface-heading">
                        <strong>{{ group.name || $t('networkManage.unknownInterface') }}</strong>
                        <el-tag v-if="group.iface" size="small" :type="group.iface.state === 'UP' ? 'success' : 'info'">
                            {{ $t(group.iface.state === 'UP' ? 'commons.status.enable' : 'commons.status.disable') }}
                        </el-tag>
                    </div>
                    <div v-if="group.iface" class="interface-body">
                        <span v-if="group.iface.reason === 'ErrNetworkVirtualBridgeReadOnly'" class="node-caption">
                            {{ $t('networkManage.virtualBridgeReadOnly') }}
                        </span>
                        <span class="node-caption">
                            {{ $t('networkManage.linkState') }}:
                            {{
                                group.iface.linkState === 'UP'
                                    ? $t('networkManage.linkConnected')
                                    : group.iface.linkState === 'DOWN' || group.iface.linkState === 'LOWERLAYERDOWN'
                                      ? $t('networkManage.linkDisconnected')
                                      : group.iface.linkState || $t('commons.status.unknown')
                            }}
                        </span>
                        <span class="mac-address">{{ group.iface.mac || '--' }}</span>
                        <div
                            v-for="ip in group.iface.addresses"
                            :key="`${ip.address}/${ip.managed}`"
                            class="address-item"
                        >
                            <span class="address-value">{{ ip.address }}</span>
                            <span class="address-status" :class="{ pending: ip.managed && !ip.applied }">
                                {{
                                    $t(
                                        ip.managed
                                            ? ip.applied
                                                ? 'networkManage.managedApplied'
                                                : 'networkManage.managedPending'
                                            : 'networkManage.readOnly',
                                    )
                                }}
                            </span>
                        </div>
                        <el-button
                            v-permission
                            class="add-address"
                            size="small"
                            link
                            type="primary"
                            :disabled="!canManage || !group.iface.backend"
                            @click.stop="$emit('add-address', group.name)"
                        >
                            + {{ $t('networkManage.addIP') }}
                        </el-button>
                    </div>
                </div>
                <div class="route-branches" :class="{ 'has-routes': group.paths.length }">
                    <div v-if="!group.paths.length" class="no-routes">{{ $t('networkManage.noRoutes') }}</div>
                    <div
                        v-for="(path, index) in group.paths"
                        :key="`${JSON.stringify(path.route)}/${index}`"
                        class="route-path"
                    >
                        <span class="route-connector" aria-hidden="true"></span>
                        <button class="route-node" @click="$emit('select-route', path.route)">
                            <span class="route-heading">
                                <span class="node-caption">
                                    {{ path.route.family }} ·
                                    {{
                                        path.route.destination === 'default' || path.route.destination.endsWith('/0')
                                            ? $t('networkManage.defaultDestination')
                                            : $t('networkManage.destination')
                                    }}
                                </span>
                                <el-tooltip
                                    v-if="
                                        path.route.reason === 'ErrNetworkVirtualBridgeReadOnly' ||
                                        path.route.reason === 'ErrNetworkLoopbackReadOnly' ||
                                        results[JSON.stringify(path.route)]?.skipped
                                    "
                                    :content="
                                        $t(
                                            path.route.reason === 'ErrNetworkLoopbackReadOnly' ||
                                                results[JSON.stringify(path.route)]?.message ===
                                                    'ErrNetworkLoopbackReadOnly'
                                                ? 'networkManage.loopbackReadOnly'
                                                : 'networkManage.virtualBridgeReadOnlyHint',
                                        )
                                    "
                                    placement="bottom"
                                >
                                    <el-tag type="info" size="small" tabindex="0">
                                        {{ $t('networkManage.unchecked') }}
                                    </el-tag>
                                </el-tooltip>
                                <el-tooltip
                                    v-else-if="results[JSON.stringify(path.route)]"
                                    :content="results[JSON.stringify(path.route)].message"
                                    :disabled="
                                        (!results[JSON.stringify(path.route)].failed &&
                                            results[JSON.stringify(path.route)].reachable) ||
                                        !results[JSON.stringify(path.route)].message
                                    "
                                    placement="bottom"
                                    popper-class="network-route-check-tooltip"
                                >
                                    <el-tag
                                        size="small"
                                        tabindex="0"
                                        :type="
                                            !results[JSON.stringify(path.route)].failed &&
                                            results[JSON.stringify(path.route)].reachable
                                                ? 'success'
                                                : 'danger'
                                        "
                                    >
                                        {{
                                            $t(
                                                results[JSON.stringify(path.route)].failed
                                                    ? 'commons.status.failed'
                                                    : results[JSON.stringify(path.route)].reachable
                                                      ? 'networkManage.reachable'
                                                      : 'networkManage.unreachable',
                                            )
                                        }}
                                    </el-tag>
                                </el-tooltip>
                            </span>
                            <strong>{{ path.route.destination }}</strong>
                            <span class="route-summary">
                                <span>
                                    {{
                                        path.gateway
                                            ? `${$t('container.gateway')}: ${path.gateway}`
                                            : $t('networkManage.direct')
                                    }}
                                </span>
                                <span>metric {{ path.route.metric }}</span>
                            </span>
                            <span v-if="path.route.reason !== 'NetworkRouteConnectedReason'" class="route-status">
                                {{
                                    $t(
                                        path.route.reason === 'ErrNetworkVirtualBridgeReadOnly'
                                            ? 'networkManage.virtualBridgeReadOnly'
                                            : path.route.managed
                                              ? 'networkManage.managedApplied'
                                              : 'networkManage.readOnly',
                                    )
                                }}
                            </span>
                        </button>
                    </div>
                </div>
            </section>
        </div>
    </div>
</template>
<script setup lang="ts">
import { computed } from 'vue';
import { HostNetwork } from '@/api/interface/host-network';
const props = defineProps<{
    interfaces: HostNetwork.Interface[];
    routes: HostNetwork.Route[];
    search: string;
    results: Record<string, HostNetwork.Reachability>;
    canManage: boolean;
}>();
defineEmits<{
    (event: 'select-interface', name: string): void;
    (event: 'select-route', route: HostNetwork.Route): void;
    (event: 'add-address', name: string): void;
}>();
const groups = computed(() => {
    const groups = new Map<
        string,
        { name: string; iface?: HostNetwork.Interface; paths: { route: HostNetwork.Route; gateway: string }[] }
    >();
    for (const iface of props.interfaces) groups.set(iface.name, { name: iface.name, iface, paths: [] });
    for (const route of props.routes) {
        const hops = route.nextHops?.length ? route.nextHops : [{ interface: route.interface, gateway: route.gateway }];
        for (const hop of hops) {
            if (!groups.has(hop.interface)) groups.set(hop.interface, { name: hop.interface, paths: [] });
            groups.get(hop.interface)!.paths.push({ route, gateway: hop.gateway });
        }
    }
    const keyword = props.search.trim().toLowerCase();
    return [...groups.values()].filter(
        (group) =>
            group.paths.length ||
            (group.iface &&
                (!keyword ||
                    `${group.name} ${group.iface.mac} ${group.iface.addresses.map((ip) => ip.address).join(' ')}`
                        .toLowerCase()
                        .includes(keyword))),
    );
});
</script>
<style scoped>
:global(.network-route-check-tooltip) {
    max-width: 360px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
}

.network-canvas {
    overflow: auto;
    max-height: 70vh;
    min-height: 360px;
}
.network-map {
    width: 100%;
    box-sizing: border-box;
    padding: 16px;
}
.map-headings {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 2fr);
    color: var(--el-text-color-secondary);
    font-size: 12px;
    padding: 0 0 12px;
}
.map-headings span:first-child {
    padding-left: 12px;
}
.map-headings span:last-child {
    padding-left: 48px;
}
.network-group {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 2fr);
    align-items: center;
    margin-bottom: 20px;
}
.network-group:last-child {
    margin-bottom: 0;
}
.interface-node {
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color);
    border-radius: 10px;
    overflow: hidden;
    z-index: 1;
}
button {
    font: inherit;
    color: var(--el-text-color-primary);
    cursor: pointer;
}
button:focus-visible,
.interface-node:focus-visible {
    outline: 2px solid var(--el-color-primary);
    outline-offset: 2px;
}
.interface-node.selectable {
    cursor: pointer;
}
.interface-heading {
    display: flex;
    align-items: center;
    gap: 8px;
    border: 0;
    border-bottom: 1px solid var(--el-border-color-lighter);
    background: var(--el-fill-color-light);
    padding: 8px 12px;
    text-align: left;
}
.interface-heading strong {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
}
.interface-heading .el-tag {
    flex-shrink: 0;
}
.interface-body {
    display: flex;
    flex-direction: column;
    padding: 8px 12px;
}
.mac-address {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.address-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
    width: 100%;
    padding: 6px 0;
    background: transparent;
    border: 0;
    border-bottom: 1px solid var(--el-border-color-extra-light);
    text-align: left;
}
.address-value {
    font-family: monospace;
    overflow-wrap: anywhere;
}
.address-status,
.route-status {
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.add-address {
    align-self: flex-start;
    margin: 8px 0 0;
    font-size: 12px;
}
.route-branches {
    position: relative;
    min-width: 0;
    padding-left: 24px;
}
.route-branches.has-routes::before {
    content: '';
    position: absolute;
    width: 24px;
    left: 0;
    top: 50%;
    border-top: 1px solid var(--el-border-color-darker);
}
.route-path {
    position: relative;
    display: grid;
    grid-template-columns: 24px minmax(0, 1fr);
    align-items: center;
    box-sizing: border-box;
    min-height: 88px;
    padding: 6px 0;
}
.route-path::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    border-left: 1px solid var(--el-border-color-darker);
}
.route-path:first-child::before {
    top: 50%;
}
.route-path:last-child::before {
    bottom: 50%;
}
.route-connector {
    border-top: 1px solid var(--el-border-color-darker);
}
.route-node {
    display: flex;
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
    justify-content: center;
    min-height: 72px;
    border: 1px solid var(--el-border-color);
    border-radius: 8px;
    background: var(--el-bg-color);
    padding: 8px 12px;
    text-align: left;
}
.route-node strong {
    font-family: monospace;
    overflow-wrap: anywhere;
    font-size: 13px;
}
.route-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 4px 8px;
    width: 100%;
}
.route-heading .el-tag {
    flex-shrink: 0;
}
.node-caption {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.route-summary {
    display: flex;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 4px 16px;
    width: 100%;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    overflow-wrap: anywhere;
}
.pending {
    color: var(--el-color-warning);
}
.route-node:hover {
    border-color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
}
.no-routes {
    padding: 20px 24px;
    color: var(--el-text-color-placeholder);
    font-size: 13px;
}
@media (max-width: 767px) {
    .network-map {
        padding: 8px;
    }
    .map-headings {
        display: none;
    }
    .network-group {
        grid-template-columns: minmax(0, 1fr);
        gap: 12px;
    }
    .route-branches {
        padding-left: 12px;
    }
    .route-branches.has-routes::before {
        display: none;
    }
    .route-path {
        grid-template-columns: 16px minmax(0, 1fr);
    }
    .route-path:first-child::before {
        top: -12px;
    }
}
</style>
