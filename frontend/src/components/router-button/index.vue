<template>
    <div>
        <el-card class="router_card p-1 sm:p-0">
            <div class="flex w-full flex-col justify-start sm:items-center items-start sm:justify-between sm:flex-row">
                <el-radio-group v-model="activeName" @change="handleChange">
                    <el-radio-button
                        class="router_card_button"
                        :label="button.label"
                        :value="button.label"
                        v-for="(button, index) in buttonArray"
                        size="large"
                        :key="index"
                    >
                        <el-badge :value="button.count" v-if="button.count" is-dot>
                            <span>{{ button.label }}</span>
                        </el-badge>
                    </el-radio-button>
                </el-radio-group>
                <div class="flex flex-col gap-2 sm:flex-row">
                    <slot name="route-button"></slot>
                </div>
            </div>
        </el-card>
        <div v-if="currentNodeVersionMismatch" class="mt-3">
            <el-alert type="warning" :closable="false" show-icon :title="$t('setting.currentNodeVersionNotSame')" />
        </div>
        <div class="mt-3" v-if="showExpiresAt && expiresAlertVisible && productProExpires && productProExpires !== 0">
            <el-alert type="warning" @close="handleExpiresAlertClose">
                <template #title>
                    <div>
                        <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                            <span>
                                {{ $t(expiresAlertKey, [expiresInfo]) }}
                            </span>
                            <el-link
                                class="cursor-pointer"
                                style="font-size: 12px"
                                icon="Position"
                                type="primary"
                                @click="goXpack"
                            >
                                {{ $t('firewall.quickJump') }}
                            </el-link>
                        </div>
                    </div>
                </template>
            </el-alert>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { routerToNameWithQuery, routerToPathWithQuery } from '@/utils/router';
import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { hasPermissionMetaAccess, hasRouteAccess } from '@/utils/rbac';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { getSettingBaseInfo } from '@/api/modules/setting';
import { listNodes } from '@/utils/node';

defineOptions({ name: 'RouterButton' });

const props = defineProps({
    buttons: {
        type: Array<RouterButton>,
        required: true,
    },
    showExpiresAt: {
        type: Boolean,
        default: false,
    },
});

const router = useRouter();
const { currentNode, isEnterprise, isIntl, isXpackOrEE, productProExpires } = useGlobalStore();
const buttonArray = computed(() => {
    return props.buttons.filter((button) => {
        if (!hasPermissionMetaAccess(button.permission)) {
            return false;
        }
        if (button.path || button.name) {
            const route = router.resolve(button.path ? { path: button.path } : { name: button.name });
            return route.matched.length === 0 || hasRouteAccess(route);
        }
        return true;
    });
});

const activeName = ref('');
const currentNodeVersionMismatch = ref(false);
const expiresInfo = ref(0);
const expiresAlertVisible = ref(false);
const expiresAlertKey = computed(() => (isEnterprise.value ? 'xpack.expiresEnterpriseAlert' : 'xpack.expiresProAlert'));

const handleChange = (label: string) => {
    const btn = buttonArray.value.find((btn) => btn.label === label);
    if (!btn) return;
    if (btn.path) routerToPathWithQuery(btn.path, { uncached: 'true' });
    else if (btn.name) routerToNameWithQuery(btn.name, { uncached: 'true' });
    activeName.value = btn.label;
};

onMounted(() => {
    syncActiveName();
    loadExpiresAlert();
});

watch(
    () => [router.currentRoute.value.path, buttonArray.value.map((button) => button.label).join('|')],
    () => {
        syncActiveName();
    },
);

watch(
    () => [props.showExpiresAt, productProExpires.value],
    () => {
        loadExpiresAlert();
    },
);

watch([currentNode, isXpackOrEE], checkCurrentNodeVersion, { immediate: true });

async function checkCurrentNodeVersion() {
    const checkedNode = currentNode.value;
    if (checkedNode === 'local' || !isXpackOrEE.value) {
        currentNodeVersionMismatch.value = false;
        return;
    }

    try {
        const [settingRes, nodes] = await Promise.all([getSettingBaseInfo(), listNodes('all')]);
        if (currentNode.value !== checkedNode) {
            return;
        }
        const currentNodeInfo = nodes.find((item) => item.name === checkedNode);
        currentNodeVersionMismatch.value = Boolean(
            currentNodeInfo?.version &&
            settingRes.data.systemVersion &&
            currentNodeInfo.version !== settingRes.data.systemVersion,
        );
    } catch {
        if (currentNode.value === checkedNode) {
            currentNodeVersionMismatch.value = false;
        }
    }
}

function syncActiveName() {
    if (!buttonArray.value.length) {
        activeName.value = '';
        return;
    }
    if (buttonArray.value.length) {
        let isPathExist = false;
        const btn = buttonArray.value.find((btn) => {
            return btn.path && router.currentRoute.value.path.startsWith(btn.path);
        });
        if (btn) {
            isPathExist = true;
            activeName.value = btn.label;
        }
        if (!isPathExist) {
            activeName.value = buttonArray.value[0].label;
        }
    }
}

function getExpiresAlertDateKey() {
    const newDate = new Date();
    return newDate.getFullYear() + '-' + newDate.getMonth() + '-' + newDate.getDate();
}

function loadExpiresAlert() {
    const expires = productProExpires.value;
    if (!props.showExpiresAt || !expires || expires === 0) {
        expiresInfo.value = 0;
        expiresAlertVisible.value = false;
        return;
    }

    if (getExpiresAlertDateKey() === localStorage.getItem('xpack-expires-alert')) {
        expiresInfo.value = 0;
        expiresAlertVisible.value = false;
        return;
    }

    const currentTimestamp = Date.now() / 1000;
    if (expires < currentTimestamp) {
        expiresInfo.value = 0;
        expiresAlertVisible.value = false;
        return;
    }

    const daySeconds = 24 * 60 * 60;
    const diffSeconds = Math.abs(expires - currentTimestamp);
    expiresInfo.value = Math.floor(diffSeconds / daySeconds) + 1;
    expiresAlertVisible.value = expiresInfo.value <= 15;
}

function goXpack() {
    if (isIntl.value && !isEnterprise.value) {
        window.open('https://1panel.hk/pricing', '_blank', 'noopener,noreferrer');
        return;
    }
    const url = isEnterprise.value ? 'https://1panel.cn/enterprise.html' : 'https://www.lxware.cn/1panel';
    window.open(url, '_blank', 'noopener,noreferrer');
}

function handleExpiresAlertClose() {
    localStorage.setItem('xpack-expires-alert', getExpiresAlertDateKey());
    loadExpiresAlert();
}
</script>

<style lang="scss" scoped>
.router_card {
    --el-card-padding: 0;

    :deep(.el-card__body) {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
}

.router_card_button {
    :deep(.el-radio-button__inner) {
        min-width: 100px;
        height: 100%;
        background-color: var(--panel-button-active) !important;
        box-shadow: none !important;
        outline: none !important;
        border: 2px solid transparent !important;
        color: var(--el-text-color-regular) !important;
    }

    :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
        color: var(--panel-button-text-color) !important;
        background-color: var(--panel-button-bg-color) !important;
        border-color: var(--panel-color-primary) !important;
        border-radius: 4px;
    }
}
</style>
