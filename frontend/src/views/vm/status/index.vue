<template>
    <div v-loading="loading" class="vm-environment-check">
        <div v-if="environmentLoaded" class="app-status card-interval">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag effect="dark" type="success">1Panel-kvm</el-tag>
                    </div>
                    <div v-if="environmentBlocker !== 'helper'" class="mt-0.5">
                        <el-divider direction="vertical" />
                        <el-button
                            type="primary"
                            link
                            :disabled="!hasEnvironmentDetails"
                            @click="dependencyVisible = true"
                        >
                            <el-icon v-if="environmentBlocker" class="vm-dependency-warning">
                                <WarningFilled />
                            </el-icon>
                            {{ $t('xpack.vm.dependencies') }}
                        </el-button>
                    </div>
                </div>
            </el-card>
        </div>

        <el-card v-if="maskMessage" class="mask-prompt vm-environment-prompt">
            <span>{{ maskMessage }}</span>
        </el-card>

        <div :class="{ mask: vmUnavailable }">
            <slot
                :vm-unavailable="vmUnavailable"
                :vm-environment-loaded="environmentLoaded"
                :refresh-vm-environment="loadEnvironment"
            />
        </div>

        <DialogPro v-model="dependencyVisible" :title="$t('xpack.vm.dependencies')" size="default">
            <el-alert
                v-if="environmentBlocker === 'architecture'"
                class="vm-environment-helper"
                type="warning"
                :closable="false"
                show-icon
                :title="$t('xpack.vm.architectureUnsupported', [environmentStatus.arch])"
            />
            <el-alert
                v-else-if="environmentBlocker === 'helper'"
                class="vm-environment-helper"
                type="warning"
                :closable="false"
                show-icon
                :title="$t('xpack.vm.helperUnavailable')"
            />
            <el-alert
                v-else-if="environmentBlocker === 'kvm'"
                class="vm-environment-helper"
                type="warning"
                :closable="false"
                show-icon
                :title="$t('xpack.vm.kvmUnavailable')"
            >
                <template #default>
                    <ol class="vm-kvm-helper__list">
                        <li>{{ $t('xpack.vm.kvmCheckCpu') }}</li>
                        <li>{{ $t('xpack.vm.kvmCheckBios') }}</li>
                        <li>{{ $t('xpack.vm.kvmCheckModule') }}</li>
                        <li>{{ $t('xpack.vm.kvmCheckDevice') }}</li>
                    </ol>
                </template>
            </el-alert>
            <template v-else>
                <el-alert
                    v-if="scriptInstallDependencyItems.length"
                    class="vm-script-library-helper"
                    type="warning"
                    :closable="false"
                    show-icon
                >
                    <template #title>
                        <span class="vm-script-library-helper__title">
                            <span>{{ $t('cronjob.library.noSuchApp', [scriptInstallDependencyNames]) }}</span>
                        </span>
                        <span class="vm-script-library-helper__link" @click="goScriptLibrary">
                            <el-icon>
                                <Position />
                            </el-icon>
                            <span>{{ $t('firewall.quickJump') }}</span>
                        </span>
                    </template>
                </el-alert>
                <ComplexTable :data="installableDependencyItems">
                    <el-table-column prop="name" :label="$t('commons.table.name')" min-width="140" />
                    <el-table-column :label="$t('commons.table.status')" min-width="120">
                        <template #default="{ row }">
                            <Status :status="row.ready ? 'enable' : 'disable'" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('xpack.vm.dependencyPurpose')" min-width="260" show-overflow-tooltip>
                        <template #default="{ row }">{{ dependencyPurpose(row) }}</template>
                    </el-table-column>
                    <el-table-column prop="message" :label="$t('commons.table.description')" show-overflow-tooltip>
                        <template #default="{ row }">{{ dependencyMessage(row) }}</template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" width="120">
                        <template #default="{ row }">
                            <el-button
                                v-if="canEnableDependency(row)"
                                v-permission
                                link
                                type="primary"
                                :loading="environmentEnabling"
                                @click="enableEnvironment"
                            >
                                {{ $t('commons.button.enable') }}
                            </el-button>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                </ComplexTable>
                <el-empty v-if="!installableDependencyItems.length" :description="$t('xpack.vm.noDependency')" />
            </template>
        </DialogPro>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { Position, WarningFilled } from '@element-plus/icons-vue';
import i18n from '@/lang';
import { VM } from '@/api/interface/vm';
import { checkVMEnvironment, enableVMEnvironment } from '@/api/modules/vm';
import { routerToNameWithQuery } from '@/utils/router';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const environmentEnabling = ref(false);
const dependencyVisible = ref(false);
const environmentLoaded = ref(false);

interface DependencyItem {
    name: string;
    ready: boolean;
    exist: boolean;
    message: string;
}

type EnvironmentBlocker = 'architecture' | 'helper' | 'kvm' | 'dependency' | '';

const defaultEnvironment = (): VM.Environment => ({
    arch: '',
    archSupported: true,
    ready: false,
    helperExist: false,
    helperMessage: '',
    kvmExist: false,
    kvmMessage: '',
    libvirtExist: false,
    libvirtActive: false,
    libvirtMessage: '',
    storageExist: false,
    storageMessage: '',
});

const environmentStatus = ref<VM.Environment>(defaultEnvironment());

const architectureUnavailable = computed(() => environmentLoaded.value && !environmentStatus.value.archSupported);
const helperUnavailable = computed(() => environmentLoaded.value && !environmentStatus.value.helperExist);
const kvmUnavailable = computed(
    () =>
        environmentLoaded.value &&
        !architectureUnavailable.value &&
        !helperUnavailable.value &&
        !environmentStatus.value.kvmExist,
);
const installableDependencyItems = computed<DependencyItem[]>(() => [
    {
        name: 'libvirt',
        ready: environmentStatus.value.libvirtActive,
        exist: environmentStatus.value.libvirtExist || environmentStatus.value.libvirtActive,
        message: environmentStatus.value.libvirtMessage,
    },
    {
        name: 'storage',
        ready: environmentStatus.value.storageExist,
        exist: environmentStatus.value.storageExist,
        message: environmentStatus.value.storageMessage,
    },
]);
const installableDependenciesIncomplete = computed(() => installableDependencyItems.value.some((item) => !item.ready));
const scriptInstallDependencyItems = computed(() =>
    installableDependencyItems.value.filter((item) => !item.ready && !item.exist),
);
const scriptInstallDependencyNames = computed(() =>
    scriptInstallDependencyItems.value.map((item) => item.name).join(', '),
);
const environmentBlocker = computed<EnvironmentBlocker>(() => {
    if (!environmentLoaded.value) {
        return '';
    }
    if (architectureUnavailable.value) {
        return 'architecture';
    }
    if (helperUnavailable.value) {
        return 'helper';
    }
    if (kvmUnavailable.value) {
        return 'kvm';
    }
    if (installableDependenciesIncomplete.value) {
        return 'dependency';
    }
    return '';
});
const hasEnvironmentDetails = computed(
    () => environmentLoaded.value && (!!environmentBlocker.value || !!installableDependencyItems.value.length),
);
const environmentReady = computed(() => environmentLoaded.value && environmentStatus.value.ready);
const vmUnavailable = computed(() => !environmentLoaded.value || !environmentReady.value || !!environmentBlocker.value);
const maskMessage = computed(() => {
    switch (environmentBlocker.value) {
        case 'architecture':
            return i18n.global.t('xpack.vm.architectureUnsupported', [environmentStatus.value.arch]);
        case 'helper':
            return i18n.global.t('xpack.vm.helperUnavailable');
        case 'kvm':
            return i18n.global.t('xpack.vm.kvmUnavailable');
        case 'dependency':
            return i18n.global.t('xpack.vm.dependencyNotReady');
        default:
            return '';
    }
});

const dependencyPurpose = (item: DependencyItem) => {
    const key = item.name.toLowerCase();
    const purposeMap: Record<string, string> = {
        libvirt: 'xpack.vm.dependencyPurposeMap.libvirt',
        storage: 'xpack.vm.dependencyPurposeMap.storage',
    };
    const i18nKey = purposeMap[key];
    return i18nKey ? i18n.global.t(i18nKey) : '-';
};

const dependencyMessage = (item: DependencyItem) => {
    if (item.message) {
        return item.message;
    }
    if (!item.exist) {
        return i18n.global.t('cronjob.library.noSuchApp', [item.name]);
    }
    return '-';
};

const canEnableDependency = (item: DependencyItem) => {
    return item.name === 'libvirt' && item.exist && !item.ready;
};

const enableEnvironment = async () => {
    environmentEnabling.value = true;
    try {
        await enableVMEnvironment();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await loadEnvironment();
    } finally {
        environmentEnabling.value = false;
    }
};

const goScriptLibrary = () => {
    dependencyVisible.value = false;
    routerToNameWithQuery('Library', { uncached: 'true' });
};

const loadEnvironment = async () => {
    loading.value = true;
    try {
        const res = await checkVMEnvironment();
        environmentStatus.value = res.data || defaultEnvironment();
    } catch {
        environmentStatus.value = defaultEnvironment();
    } finally {
        environmentLoaded.value = true;
        dependencyVisible.value = false;
        loading.value = false;
    }
};

onMounted(loadEnvironment);
</script>

<style scoped lang="scss">
.vm-environment-check {
    position: relative;
    min-height: 240px;
}

.vm-environment-prompt {
    z-index: 2;
}

.vm-dependency-warning {
    margin-right: 4px;
    color: var(--el-color-warning);
    vertical-align: -1px;
}

.vm-environment-helper {
    margin-bottom: 12px;
}

.vm-script-library-helper {
    margin-bottom: 12px;
}

.vm-script-library-helper__title {
    display: inline;
    line-height: 18px;
}

.vm-script-library-helper__link {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    margin-left: 8px;
    line-height: 18px;
    color: var(--el-color-primary);
    cursor: pointer;
    white-space: nowrap;
}

.vm-kvm-helper__list {
    margin: 6px 0 0 0;
    padding-left: 18px;
    line-height: 22px;
}
</style>
