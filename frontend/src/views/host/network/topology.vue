<template>
    <LayoutContent v-loading="loading" :title="$t('networkManage.title')">
        <template #leftToolBar>
            <el-button
                v-permission
                type="primary"
                :disabled="!canManage || busy || !interfaces.some((item) => item.backend)"
                @click="openRoute()"
            >
                {{ $t('networkManage.addRoute') }}
            </el-button>
            <el-button
                :loading="checking"
                :disabled="!ready || busy || !filteredRoutes.some(canCheckRoute)"
                @click="checkRoutes(filteredRoutes)"
            >
                {{ $t('networkManage.checkRoutes') }}
            </el-button>
        </template>
        <template #rightToolBar>
            <TableSearch
                v-model:searchName="search"
                class="network-search"
                :placeholder="$t('networkManage.topologySearch')"
            />
            <TableRefresh :disabled="busy" @search="load" />
            <TableSetting title="host-network-topology-refresh" @search="load" />
        </template>
        <template #main>
            <p class="topology-hint">{{ $t('networkManage.topologyHint') }}</p>
            <NetworkGraph
                v-if="ready"
                :interfaces="interfaces"
                :routes="filteredRoutes"
                :search="search"
                :results="results"
                :can-manage="canManage && !busy"
                @select-interface="detail = { kind: 'interface', name: $event }"
                @select-route="detail = { kind: 'route', key: routeKey($event) }"
                @add-address="openAddress"
            />
        </template>
    </LayoutContent>
    <DrawerPro
        :model-value="!!detail"
        :header="$t(detail?.kind === 'route' ? 'networkManage.routeDetails' : 'networkManage.interfaces')"
        :size="detail?.kind === 'interface' ? '50%' : 'normal'"
        @update:model-value="!$event && (detail = undefined)"
    >
        <template v-if="selectedInterface">
            <el-descriptions :title="selectedInterface.name" :column="2" border class="mt-3 sm:mt-0">
                <el-descriptions-item :label="$t('networkManage.persistenceBackend')">
                    {{ selectedInterface.backend || '--' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.interfaceID')">
                    {{ selectedInterface.index || '--' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.adminState')">
                    {{ $t(selectedInterface.state === 'UP' ? 'commons.status.enable' : 'commons.status.disable') }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.linkState')">
                    {{
                        selectedInterface.linkState === 'UP'
                            ? $t('networkManage.linkConnected')
                            : selectedInterface.linkState === 'DOWN' || selectedInterface.linkState === 'LOWERLAYERDOWN'
                              ? $t('networkManage.linkDisconnected')
                              : selectedInterface.linkState || $t('commons.status.unknown')
                    }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.mac')">
                    {{ selectedInterface.mac || '--' }}
                </el-descriptions-item>
                <el-descriptions-item label="MTU">{{ selectedInterface.mtu || '--' }}</el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.ipCount')">
                    {{ selectedInterface.addresses.length }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.flags')" :span="2">
                    {{ selectedInterface.flags.join(', ') || '--' }}
                </el-descriptions-item>
            </el-descriptions>
            <fu-table :data="selectedInterface.addresses" class="mt-5" row-key="address">
                <el-table-column :label="$t('networkManage.address')" prop="address" min-width="180" />
                <el-table-column :label="$t('networkManage.scope')" prop="scope" width="100" />
                <el-table-column :label="$t('networkManage.configStatus')" min-width="145">
                    <template #default="{ row }">
                        <el-tooltip
                            :content="networkReason(row.reason, selectedInterface.name)"
                            :disabled="row.managed || !row.reason"
                            placement="bottom"
                        >
                            <el-tag :type="row.managed ? (row.applied ? 'success' : 'warning') : 'info'">
                                {{
                                    $t(
                                        row.managed
                                            ? row.applied
                                                ? 'networkManage.managedApplied'
                                                : 'networkManage.managedPending'
                                            : 'networkManage.readOnly',
                                    )
                                }}
                            </el-tag>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="150" fixed="right">
                    <template #default="{ row }">
                        <el-button
                            v-permission
                            type="primary"
                            link
                            :disabled="!row.editable || !canManage || busy"
                            @click="openAddress(selectedInterface.name, row)"
                        >
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button
                            v-permission
                            type="primary"
                            link
                            :disabled="!row.editable || !canManage || busy"
                            @click="removeAddress(row)"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </fu-table>

            <el-button
                v-permission
                type="primary"
                :disabled="!canManage || busy || !selectedInterface.backend"
                @click="openAddress(selectedInterface.name)"
            >
                {{ $t('networkManage.addIP') }}
            </el-button>
        </template>
        <template v-if="selectedRoute">
            <el-alert
                v-if="selectedRoute.reason === 'NetworkRouteConnectedReason'"
                class="connected-route-alert"
                type="info"
                show-icon
                :closable="false"
                :title="$t('networkManage.connectedRouteHint')"
            />
            <el-descriptions :column="1" border>
                <el-descriptions-item :label="$t('networkManage.destination')">
                    {{ selectedRoute.destination }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('container.gateway')">
                    {{ selectedRoute.gateway || '--' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.interface')">
                    {{ selectedRoute.interface || '--' }}
                </el-descriptions-item>
                <el-descriptions-item v-if="selectedRoute.nextHops?.length" :label="$t('networkManage.routes')">
                    <div v-for="(hop, index) in selectedRoute.nextHops" :key="index">
                        {{ hop.interface || '--' }} → {{ hop.gateway || $t('networkManage.direct') }} ({{ hop.weight }})
                    </div>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('commons.table.type')">
                    {{ selectedRoute.family }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.metric')">
                    {{ selectedRoute.metric }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('commons.table.protocol')">
                    {{ selectedRoute.protocol || '--' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.scope')">
                    {{ selectedRoute.scope || '--' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.sourceAddress')">
                    {{ selectedRoute.source || '--' }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.routeTable')">
                    {{ selectedRoute.table }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.configStatus')">
                    {{
                        $t(
                            selectedRoute.reason === 'ErrNetworkVirtualBridgeReadOnly'
                                ? 'networkManage.virtualBridgeReadOnly'
                                : selectedRoute.reason === 'NetworkRouteConnectedReason'
                                  ? 'networkManage.connectedRoute'
                                  : selectedRoute.managed
                                    ? 'networkManage.managedApplied'
                                    : 'networkManage.readOnly',
                        )
                    }}
                    <p
                        v-if="
                            !selectedRoute.managed &&
                            selectedRoute.reason &&
                            selectedRoute.reason !== 'NetworkRouteConnectedReason'
                        "
                        class="readonly-reason"
                    >
                        {{ networkReason(selectedRoute.reason, selectedRoute.interface) }}
                    </p>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('networkManage.reachability')">
                    <template v-if="selectedRouteCheck">
                        <el-tag
                            :type="
                                selectedRouteCheck.skipped
                                    ? 'info'
                                    : !selectedRouteCheck.failed && selectedRouteCheck.reachable
                                      ? 'success'
                                      : 'danger'
                            "
                        >
                            {{
                                $t(
                                    selectedRouteCheck.skipped
                                        ? 'networkManage.unchecked'
                                        : selectedRouteCheck.failed
                                          ? 'commons.status.failed'
                                          : selectedRouteCheck.reachable
                                            ? 'networkManage.reachable'
                                            : 'networkManage.unreachable',
                                )
                            }}
                        </el-tag>
                        <p
                            v-if="
                                !selectedRouteCheck.skipped &&
                                (selectedRouteCheck.failed || !selectedRouteCheck.reachable) &&
                                selectedRouteCheck.message
                            "
                            class="route-check-message"
                        >
                            {{ networkReason(selectedRouteCheck.message, selectedRoute.interface) }}
                        </p>
                    </template>
                    <template v-else>{{ $t('networkManage.unchecked') }}</template>
                </el-descriptions-item>
            </el-descriptions>
            <el-alert
                v-if="
                    selectedRoute.readOnly &&
                    selectedRoute.reason !== 'ErrNetworkVirtualBridgeReadOnly' &&
                    selectedRoute.reason !== 'ErrNetworkLoopbackReadOnly'
                "
                class="mt-4"
                type="info"
                :closable="false"
                :title="$t('networkManage.readOnlyRouteHint')"
            />
            <div class="mt-4">
                <el-button
                    v-if="selectedRoute.reason !== 'NetworkRouteConnectedReason'"
                    v-permission
                    type="primary"
                    :disabled="!canManage || busy || !selectedRoute.editable"
                    @click="openRoute(selectedRoute)"
                >
                    {{ $t('commons.button.edit') }}
                </el-button>
                <el-button
                    :disabled="!ready || busy || !canCheckRoute(selectedRoute)"
                    :loading="checking"
                    @click="checkRoutes([selectedRoute])"
                >
                    {{ $t('networkManage.check') }}
                </el-button>
                <el-button
                    v-if="selectedRoute.reason !== 'NetworkRouteConnectedReason'"
                    v-permission
                    :disabled="!canManage || busy || selectedRoute.readOnly || !selectedRoute.editable"
                    @click="removeRoute(selectedRoute)"
                >
                    {{ $t('commons.button.delete') }}
                </el-button>
            </div>
        </template>
    </DrawerPro>
    <DrawerPro
        v-model="drawer"
        :header="$t(previousRoute ? 'networkManage.editRoute' : 'networkManage.addRoute')"
        size="large"
    >
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="saving">
            <el-form-item :label="$t('networkManage.destination')" prop="destination">
                <el-input
                    v-model.trim="form.destination"
                    placeholder="10.0.0.0/24, 10.0.0.5, 2001:db8::/64, 2001:db8::1"
                />
                <span class="input-help">{{ $t('networkManage.destinationHint') }}</span>
            </el-form-item>
            <el-form-item :label="$t('container.gateway')" prop="gateway">
                <el-input v-model.trim="form.gateway" />
                <span class="input-help w-full">{{ $t('networkManage.gatewayHint') }}</span>
            </el-form-item>
            <el-form-item :label="$t('networkManage.interface')" prop="interface">
                <el-select v-model="form.interface" :disabled="!!previousRoute" filterable class="w-full">
                    <el-option
                        v-for="item in interfaces.filter((item) => item.backend)"
                        :key="item.name"
                        :label="`${item.name} · ${$t(item.state === 'UP' ? 'commons.status.enable' : 'commons.status.disable')} / ${item.linkState === 'UP' ? $t('networkManage.linkConnected') : item.linkState === 'DOWN' || item.linkState === 'LOWERLAYERDOWN' ? $t('networkManage.linkDisconnected') : item.linkState || $t('commons.status.unknown')}`"
                        :value="item.name"
                    />
                </el-select>
                <span class="input-help w-full">{{ $t('networkManage.interfaceSelectHint') }}</span>
            </el-form-item>
            <el-form-item :label="$t('networkManage.metric')" prop="metric">
                <el-input-number v-model="form.metric" :min="0" :max="4294967295" :precision="0" />
                <span class="input-help w-full">{{ $t('networkManage.metricHint') }}</span>
            </el-form-item>
            <el-alert
                v-if="isDefaultRoute(form.destination)"
                :title="$t('networkManage.defaultRouteWarning')"
                type="warning"
                :closable="false"
            />
        </el-form>
        <template #footer>
            <el-button :disabled="saving" @click="drawer = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" :loading="saving" :disabled="!canManage" @click="saveRoute">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <NetworkChangeDialog
        :model-value="!!pendingRoute"
        :title="$t(pendingRoute?.previous ? 'networkManage.editRoute' : 'networkManage.addRoute')"
        :node="nodeDisplayName"
        :device="pendingRoute?.interface || ''"
        :loading="saving"
        :disabled="!canManage || busy"
        @update:model-value="!$event && (pendingRoute = undefined)"
        @confirm="confirmRoute"
    >
        <template v-if="pendingRoute && isDefaultRoute(pendingRoute.destination)" #warning>
            <div>{{ $t('networkManage.defaultRouteWarning') }}</div>
        </template>
        <template v-if="pendingRoute" #details>
            <el-descriptions-item :label="$t('networkManage.destination')">
                <template v-if="pendingRoute.previous">{{ pendingRoute.previous.destination }} →</template>
                <code>{{ pendingRoute.destination }}</code>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('container.gateway')">
                <template v-if="pendingRoute.previous">{{ pendingRoute.previous.gateway || '--' }} →</template>
                {{ pendingRoute.gateway || '--' }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('networkManage.metric')">
                <template v-if="pendingRoute.previous">{{ pendingRoute.previous.metric }} →</template>
                {{ pendingRoute.metric }}
            </el-descriptions-item>
        </template>
    </NetworkChangeDialog>
    <NetworkChangeDialog
        v-model="addressDialog"
        :title="$t(previousAddress ? 'networkManage.editIP' : 'networkManage.addIP')"
        :node="nodeDisplayName"
        :device="addressForm.interface"
        :loading="saving"
        :disabled="!canManage || busy"
        @confirm="saveAddress"
    >
        <template #warning>
            <div>{{ $t('networkManage.addressWarning') }}</div>
        </template>
        <template #details>
            <el-descriptions-item v-if="previousAddress" :label="$t('networkManage.address')">
                <code>{{ previousAddress }}</code>
            </el-descriptions-item>
        </template>
        <el-form
            ref="addressFormRef"
            :model="addressForm"
            :rules="addressRules"
            :disabled="saving"
            label-position="top"
            @submit.prevent
            @keydown.enter.prevent.stop
        >
            <el-form-item :label="$t('networkManage.addressWithPrefix')" prop="address">
                <el-input v-model.trim="addressForm.address" placeholder="192.0.2.10/24" />
                <span class="input-help">{{ $t('networkManage.cidrHint') }}</span>
            </el-form-item>
        </el-form>
    </NetworkChangeDialog>
    <NetworkChangeDialog
        :model-value="!!deleteTarget"
        :title="$t('commons.button.delete')"
        :node="nodeDisplayName"
        :device="deleteTarget?.params.interface || ''"
        :loading="saving"
        :disabled="!canManage || busy"
        confirm-type="danger"
        :confirm-text="$t('commons.button.delete')"
        @update:model-value="!$event && (deleteTarget = undefined)"
        @confirm="confirmDelete"
    >
        <template #details>
            <template v-if="deleteTarget">
                <el-descriptions-item v-if="deleteTarget.kind === 'address'" :label="$t('networkManage.address')">
                    <code>{{ deleteTarget.params.address }}</code>
                </el-descriptions-item>
                <template v-else>
                    <el-descriptions-item :label="$t('networkManage.destination')">
                        <code>{{ deleteTarget.params.destination }}</code>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('container.gateway')">
                        {{ deleteTarget.params.gateway || '--' }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('networkManage.metric')">
                        {{ deleteTarget.params.metric }}
                    </el-descriptions-item>
                </template>
            </template>
        </template>
    </NetworkChangeDialog>
</template>
<script lang="ts" setup>
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref, watch } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import i18n from '@/lang';
import { HostNetwork } from '@/api/interface/host-network';
import {
    addHostNetworkAddress,
    deleteHostNetworkAddress,
    addHostNetworkRoute,
    checkHostNetworkRoute,
    checkHostNetworkRoutes,
    deleteHostNetworkRoute,
    getHostNetworkOverview,
} from '@/api/modules/host-network';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import NetworkGraph from './network-graph.vue';
import NetworkChangeDialog from './network-change-dialog.vue';
import { validNetworkCIDR, validNetworkIP } from './validation';

const networkReason = (reason: string, device: string) => {
    switch (reason) {
        case 'ErrNetworkVirtualBridgeReadOnly':
            return i18n.global.t('networkManage.virtualBridgeReadOnlyHint');
        case 'ErrNetworkLoopbackReadOnly':
            return i18n.global.t('networkManage.loopbackReadOnly');
        case 'ErrNetworkPersistentSourceMissing':
            return i18n.global.t('networkManage.persistentSourceMissing', [device]);
        case 'NetworkAddressReadOnlyReason':
            return i18n.global.t('networkManage.addressReadOnlyReason');
        case 'NetworkRouteConnectedReason':
            return i18n.global.t('networkManage.connectedRouteHint');
        case 'NetworkRouteReadOnlyReason':
            return i18n.global.t('networkManage.routeReadOnlyReason');
        default:
            return reason;
    }
};

const { currentNode, globalStore } = useGlobalStore();
const dataNode = ref('');
const nodeDisplayName = computed(() => (dataNode.value === 'local' ? globalStore.getMasterAlias() : dataNode.value));
const ready = computed(() => !!dataNode.value && dataNode.value === currentNode.value);
const persistence = ref<HostNetwork.Persistence>();
const canManage = computed(() => ready.value && persistence.value?.supported === true);
let active = false;
let generation = 0;
let controller: AbortController | undefined;

const loading = ref(false);
const saving = ref(false);
const checking = ref(false);
const drawer = ref(false);
const pendingRoute = ref<HostNetwork.RouteChange>();
const search = ref('');
const routes = ref<HostNetwork.Route[]>([]);
const interfaces = ref<HostNetwork.Interface[]>([]);
const results = ref<Record<string, HostNetwork.Reachability>>({});
const detail = ref<{ kind: 'interface'; name: string } | { kind: 'route'; key: string }>();
const selectedInterface = computed(() => {
    const selection = detail.value;
    return selection?.kind === 'interface' ? interfaces.value.find((item) => item.name === selection.name) : undefined;
});
const selectedRoute = computed(() => {
    const selection = detail.value;
    return selection?.kind === 'route' ? routes.value.find((item) => routeKey(item) === selection.key) : undefined;
});
const selectedRouteCheck = computed<HostNetwork.Reachability | undefined>(() => {
    const route = selectedRoute.value;
    if (route?.reason === 'ErrNetworkVirtualBridgeReadOnly' || route?.reason === 'ErrNetworkLoopbackReadOnly') {
        return { skipped: true, reachable: false, message: route.reason };
    }
    return route ? results.value[routeKey(route)] : undefined;
});
const deleteTarget = ref<
    { kind: 'address'; params: HostNetwork.AddressRequest } | { kind: 'route'; params: HostNetwork.Route }
>();
const addressDialog = ref(false);
const previousAddress = ref('');
const previousRoute = ref<HostNetwork.Route>();
const addressFormRef = ref<FormInstance>();
const addressForm = reactive<HostNetwork.AddressRequest>({ interface: '', address: '' });
const addressRules: FormRules = {
    address: [
        {
            required: true,
            validator: (_rule, value, callback) =>
                callback(validNetworkCIDR(value) ? undefined : new Error(i18n.global.t('commons.rule.formatErr'))),
            trigger: 'blur',
        },
    ],
};
const busy = computed(() => loading.value || saving.value || checking.value);
const filteredRoutes = computed(() => {
    const keyword = search.value.trim().toLowerCase();
    return routes.value.filter(
        (item) =>
            !keyword ||
            `${item.destination} ${item.gateway} ${item.interface} ${item.nextHops?.map((hop) => `${hop.gateway} ${hop.interface}`).join(' ')}`
                .toLowerCase()
                .includes(keyword) ||
            interfaces.value.some(
                (iface) =>
                    (iface.name === item.interface || item.nextHops?.some((hop) => hop.interface === iface.name)) &&
                    `${iface.name} ${iface.mac} ${iface.addresses.map((ip) => ip.address).join(' ')}`
                        .toLowerCase()
                        .includes(keyword),
            ),
    );
});
const routeKey = (row: HostNetwork.Route) => JSON.stringify(row);
const canCheckRoute = (row: HostNetwork.Route) => !row.readOnly;
const isDefaultRoute = (destination: string) => destination === 'default' || destination.endsWith('/0');
const formRef = ref<FormInstance>();
const initialRoute = (): HostNetwork.Route => ({
    editable: false,
    reason: '',
    destination: '',
    gateway: '',
    interface: '',
    metric: 0,
    table: 'main',
    protocol: 'static',
    scope: '',
    family: 'IPv4',
    readOnly: false,
    nextHops: [],
    managed: false,
    applied: false,
});
const form = reactive(initialRoute());
const routeFamily = computed(() =>
    (form.destination === 'default' ? form.gateway : form.destination).includes(':') ? 'IPv6' : 'IPv4',
);
const rules: FormRules = {
    destination: [
        {
            required: true,
            validator: (_rule, value, callback) =>
                callback(
                    value === 'default' || validNetworkCIDR(value) || validNetworkIP(value)
                        ? undefined
                        : new Error(i18n.global.t('commons.rule.formatErr')),
                ),
            trigger: 'blur',
        },
    ],
    gateway: [
        {
            validator: (_rule, value, callback) =>
                callback(
                    !value || (validNetworkIP(value) && value.includes(':') === (routeFamily.value === 'IPv6'))
                        ? undefined
                        : new Error(i18n.global.t('commons.rule.formatErr')),
                ),
            trigger: 'blur',
        },
    ],
    interface: [{ required: true, message: i18n.global.t('commons.rule.requiredSelect'), trigger: 'change' }],
    metric: [
        {
            type: 'number',
            required: true,
            min: 0,
            max: 4294967295,
            message: i18n.global.t('commons.rule.formatErr'),
            trigger: 'blur',
        },
    ],
};

const isCurrent = (node: string, version: number) => active && version === generation && node === currentNode.value;
const reset = () => {
    generation++;
    controller?.abort();
    dataNode.value = '';
    persistence.value = undefined;
    routes.value = [];
    interfaces.value = [];
    results.value = {};
    drawer.value = false;
    pendingRoute.value = undefined;
    addressDialog.value = false;
    deleteTarget.value = undefined;
    detail.value = undefined;
    loading.value = false;
    saving.value = false;
    checking.value = false;
};
const load = async () => {
    if (!active || loading.value || checking.value || saving.value || pendingRoute.value) return;
    const node = currentNode.value;
    const version = ++generation;
    controller?.abort();
    controller = new AbortController();
    dataNode.value = '';
    loading.value = true;
    try {
        const { data } = await getHostNetworkOverview(node, controller.signal);
        if (!isCurrent(node, version)) return;
        persistence.value = data.persistence;
        interfaces.value = data.interfaces;
        routes.value = data.routes.map((route) => ({
            ...route,
            destination:
                route.destination === 'default' ? (route.family === 'IPv6' ? '::/0' : '0.0.0.0/0') : route.destination,
        }));
        dataNode.value = node;
        results.value = {};
        if (detail.value && !selectedInterface.value && !selectedRoute.value) detail.value = undefined;
    } catch {
        if (!isCurrent(node, version)) return;
        persistence.value = undefined;
        interfaces.value = [];
        routes.value = [];
        results.value = {};
        detail.value = undefined;
    } finally {
        if (isCurrent(node, version)) loading.value = false;
    }
};
const openRoute = (route?: HostNetwork.Route) => {
    if (!canManage.value || busy.value) return;
    previousRoute.value = route ? { ...route } : undefined;
    Object.assign(form, route || initialRoute(), {
        interface: route?.interface || selectedInterface.value?.name || '',
    });
    drawer.value = true;
    formRef.value?.clearValidate();
};
const openAddress = (name: string, address?: HostNetwork.Address) => {
    if (!canManage.value || busy.value || !interfaces.value.some((item) => item.name === name && item.backend)) return;
    addressForm.interface = name;
    previousAddress.value = address?.address || '';
    addressForm.address = address?.address || '';
    addressDialog.value = true;
    addressFormRef.value?.clearValidate();
};
const saveRoute = async () => {
    if (!canManage.value || busy.value || pendingRoute.value) return;
    const node = dataNode.value;
    const version = generation;
    saving.value = true;
    try {
        if (!(await formRef.value?.validate().catch(() => false))) return;
        if (!isCurrent(node, version) || !drawer.value) return;
        pendingRoute.value = {
            ...form,
            destination:
                form.destination === 'default'
                    ? routeFamily.value === 'IPv6'
                        ? '::/0'
                        : '0.0.0.0/0'
                    : form.destination,
            family: routeFamily.value,
            previous: previousRoute.value,
            confirmation: '',
        };
    } finally {
        if (isCurrent(node, version)) saving.value = false;
    }
};
const confirmRoute = async (confirmation: string) => {
    const params = pendingRoute.value;
    if (!params || !canManage.value || busy.value || confirmation !== params.interface) return;
    const node = dataNode.value;
    const version = generation;
    saving.value = true;
    try {
        await addHostNetworkRoute({ ...params, confirmation }, node);
        if (!isCurrent(node, version)) return;
        drawer.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch {
    } finally {
        if (isCurrent(node, version)) {
            saving.value = false;
            pendingRoute.value = undefined;
        }
    }
    if (isCurrent(node, version)) await load();
};
const removeRoute = (row: HostNetwork.Route) => {
    if (row.readOnly || !row.editable || !canManage.value || busy.value) return;
    deleteTarget.value = { kind: 'route', params: { ...row } };
};
const saveAddress = async (confirmation: string) => {
    if (!canManage.value || busy.value || confirmation !== addressForm.interface) return;
    const node = dataNode.value;
    const version = generation;
    const params: HostNetwork.AddressChange = {
        ...addressForm,
        previousAddress: previousAddress.value,
        confirmation,
    };
    saving.value = true;
    let refresh = false;
    try {
        if (!(await addressFormRef.value?.validate().catch(() => false))) return;
        if (!isCurrent(node, version)) return;
        refresh = true;
        await addHostNetworkAddress(params, node);
        if (!isCurrent(node, version)) return;
        addressDialog.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch {
    } finally {
        if (isCurrent(node, version)) saving.value = false;
    }
    if (refresh && isCurrent(node, version)) await load();
};
const removeAddress = (address: HostNetwork.Address) => {
    if (!address.editable || !canManage.value || busy.value || !selectedInterface.value) return;
    deleteTarget.value = {
        kind: 'address',
        params: { interface: selectedInterface.value.name, address: address.address },
    };
};
const confirmDelete = async (confirmation: string) => {
    const target = deleteTarget.value;
    if (!target || !canManage.value || busy.value || confirmation !== target.params.interface) return;
    const node = dataNode.value;
    const version = generation;
    saving.value = true;
    try {
        if (target.kind === 'address') {
            await deleteHostNetworkAddress({ ...target.params, confirmation }, node);
        } else {
            await deleteHostNetworkRoute({ ...target.params, confirmation }, node);
        }
        if (!isCurrent(node, version)) return;
        deleteTarget.value = undefined;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch {
    } finally {
        if (isCurrent(node, version)) saving.value = false;
    }
    if (isCurrent(node, version)) await load();
};
const checkRoutes = async (items: HostNetwork.Route[]) => {
    if (!ready.value || checking.value || loading.value || saving.value) return;
    const node = dataNode.value;
    const version = generation;
    controller?.abort();
    controller = new AbortController();
    const signal = controller.signal;
    checking.value = true;
    try {
        const rows = items.filter(canCheckRoute);
        for (const row of rows) {
            delete results.value[routeKey(row)];
        }
        if (rows.length === 1) {
            const res = await checkHostNetworkRoute(rows[0], node, signal);
            if (isCurrent(node, version)) {
                results.value[routeKey(rows[0])] = res.data;
            }
            return;
        }
        const batchSize = 16;
        for (let offset = 0; offset < rows.length; offset += batchSize) {
            const batch = rows.slice(offset, offset + batchSize);
            const res = await checkHostNetworkRoutes(batch, node, signal);
            if (!isCurrent(node, version)) return;
            batch.forEach((row, index) => {
                results.value[routeKey(row)] = res.data[index];
            });
        }
    } catch {
    } finally {
        if (isCurrent(node, version)) checking.value = false;
    }
};
const activate = () => {
    if (active) return;
    active = true;
    load();
};
const deactivate = () => {
    active = false;
    reset();
};
watch(
    currentNode,
    () => {
        reset();
        search.value = '';
        if (active) load();
    },
    { flush: 'sync' },
);
onMounted(activate);
onActivated(activate);
onDeactivated(deactivate);
onBeforeUnmount(deactivate);
</script>
<style scoped>
.connected-route-alert {
    margin-bottom: 16px;
}

.readonly-reason {
    margin: 8px 0 0;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
}

.route-check-message {
    margin: 8px 0 0;
    color: var(--el-color-danger);
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    line-height: 1.6;
}

.network-search {
    width: 320px;
    max-width: 100%;
}
.network-search :deep(.search-button) {
    width: 100%;
}
.topology-hint {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    margin: 0 0 12px;
    line-height: 1.6;
}
</style>
