<template>
    <DrawerPro
        v-model="visible"
        :header="$t(isBatch ? 'container.batchUpgrade' : 'commons.button.upgrade')"
        size="large"
        :confirm-before-close="true"
        @before-close="beforeClose"
        @close="handleClose"
    >
        <el-form ref="formRef" :model="form" label-position="top" @submit.prevent>
            <section v-for="(group, index) in form.groups" :key="group.currentImage" class="upgrade-group">
                <div v-if="isBatch" class="mb-3">
                    <span v-for="name in group.names" :key="name" class="group-name mr-2 mb-2">
                        {{ name }}
                    </span>
                </div>
                <el-alert
                    v-if="group.appNames.some((name) => group.names.includes(name))"
                    :title="$t('container.appHelper')"
                    type="warning"
                    :closable="false"
                    class="common-prompt app-warning"
                />
                <el-form-item :label="$t('container.oldImage')">
                    <el-tag class="image-name">{{ group.currentImage }}</el-tag>
                </el-form-item>
                <el-form-item v-if="group.options.length > 1" :label="$t('container.sameImageContainer')">
                    <div class="w-full">
                        <el-checkbox
                            :model-value="isGroupChecked(group)"
                            :indeterminate="group.names.length > 0 && !isGroupChecked(group)"
                            :disabled="loading || !isGroupEditable(group)"
                            @change="(value) => handleCheckAllChange(group, value)"
                        >
                            {{ $t('commons.table.all') }}
                        </el-checkbox>
                    </div>
                    <el-checkbox-group v-model="group.names" :disabled="loading || !isGroupEditable(group)">
                        <el-checkbox v-for="item in group.options" :key="item.name" :value="item.name">
                            {{ item.name }}
                        </el-checkbox>
                    </el-checkbox-group>
                    <span class="input-help">{{ $t('container.sameImageHelper') }}</span>
                </el-form-item>
                <el-form-item :prop="`groups.${index}.targetImage`" :rules="group.names.length ? Rules.imageName : []">
                    <template #label>
                        {{ $t('container.targetImage') }}
                        <span v-if="group.currentImage.includes('sha256:')">
                            {{ ' (' + $t('container.imageLoadErr') + ')' }}
                        </span>
                    </template>
                    <el-input
                        v-model="group.targetImage"
                        :disabled="loading || !group.names.length || !isGroupEditable(group)"
                    />
                    <span class="input-help">{{ $t('container.upgradeHelper') }}</span>
                </el-form-item>
                <div v-if="group.state !== 'draft'" class="mb-3">
                    <el-tag :type="statusType(group)">{{ statusLabel(group) }}</el-tag>
                    <el-button v-if="group.state === 'submitted'" link type="primary" @click="openLog(group)">
                        {{ $t('container.batchUpgradeLog') }}
                    </el-button>
                    <el-button v-if="group.state === 'unknown'" link type="primary" @click="refreshGroup(group)">
                        {{ $t('container.batchUpgradeCheck') }}
                    </el-button>
                </div>
            </section>
            <el-form-item>
                <el-checkbox v-model="form.forcePull" :disabled="started || loading">
                    {{ $t('container.forcePull') }}
                </el-checkbox>
                <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="visible = false">
                {{ $t('commons.button.close') }}
            </el-button>
            <el-button v-if="canSubmit" type="primary" :loading="loading" @click="submit">
                {{ started ? $t('container.batchUpgradeRetry') : $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" width="70%" @close="emit('search')" />
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, reactive, ref } from 'vue';
import { ElMessageBox, type CheckboxValueType, type FormInstance } from 'element-plus';
import { Container } from '@/api/interface/container';
import { listContainerByImage, upgradeContainer } from '@/api/modules/container';
import { searchTasks } from '@/api/modules/log';
import { Rules } from '@/global/form-rules';
import { newUUID } from '@/utils/id';
import i18n from '@/lang';
import TaskLog from '@/components/log/task/index.vue';
import bus from '@/global/bus';
import { GlobalStore } from '@/store';

interface UpgradeGroup {
    currentImage: string;
    targetImage: string;
    names: string[];
    appNames: string[];
    options: Container.ContainerOption[];
    taskID: string;
    state: 'draft' | 'submitting' | 'submitted' | 'failed' | 'unknown';
    taskStatus: string;
    currentStep: string;
}
interface DialogProps {
    container: string;
    image: string;
    fromApp: boolean;
    targetImage?: string;
    hasUpdate?: boolean;
}
const isBatch = ref(false);
const isGroupEditable = (group: UpgradeGroup) => ['draft', 'failed'].includes(group.state);
const isGroupChecked = (group: UpgradeGroup) => group.options.length > 0 && group.names.length === group.options.length;
const handleCheckAllChange = (group: UpgradeGroup, value: CheckboxValueType) => {
    if (loading.value || !isGroupEditable(group)) return;
    group.names = value ? group.options.map((item) => item.name) : [];
};
const visible = ref(false);
const loading = ref(false);
const started = ref(false);
const formRef = ref<FormInstance>();
const taskLogRef = ref<InstanceType<typeof TaskLog>>();
const form = reactive({ groups: [] as UpgradeGroup[], forcePull: false });
const emit = defineEmits<{ (e: 'search'): void }>();
let timer: ReturnType<typeof setTimeout> | undefined;
let generation = 0;
let operateNode = '';
const canSubmit = computed(() => form.groups.some((group) => group.names.length > 0 && isGroupEditable(group)));

const acceptParams = async (props: DialogProps | Container.ContainerInfo[]) => {
    if (loading.value) return;
    isBatch.value = Array.isArray(props);
    const containers = Array.isArray(props)
        ? props
        : [
              {
                  name: props.container,
                  imageName: props.image,
                  isFromApp: props.fromApp,
                  targetImage: props.targetImage,
                  hasUpdate: props.hasUpdate,
              },
          ];
    operateNode = GlobalStore().currentNode;
    clearTimeout(timer);
    generation++;
    const groups = new Map<string, UpgradeGroup>();
    for (const item of containers) {
        let group = groups.get(item.imageName);
        if (!group) {
            group = {
                currentImage: item.imageName,
                targetImage: item.imageName.includes('sha256:') ? '' : item.imageName,
                names: [],
                appNames: [],
                options: [],
                taskID: newUUID(),
                state: 'draft',
                taskStatus: '',
                currentStep: '',
            };
            groups.set(item.imageName, group);
        }
        if (item.hasUpdate && item.targetImage) group.targetImage = item.targetImage;
        if (!group.names.includes(item.name)) {
            group.names.push(item.name);
            group.options.push({ name: item.name, state: '' });
        }
        if (item.isFromApp && !group.appNames.includes(item.name)) group.appNames.push(item.name);
    }
    form.groups = [...groups.values()];
    form.forcePull = containers.some((item) => item.hasUpdate);
    started.value = false;
    visible.value = true;
    formRef.value?.clearValidate();
    const version = generation;
    const node = operateNode;
    const pending = [...form.groups];
    for (let index = 0; index < pending.length; index += 3) {
        if (version !== generation || !visible.value || started.value) return;
        await Promise.all(
            pending.slice(index, index + 3).map(async (group) => {
                try {
                    const res = await listContainerByImage(group.currentImage, node);
                    if (version !== generation || !visible.value || started.value) return;
                    const options = new Map(group.options.map((item) => [item.name, item]));
                    for (const item of res.data || []) options.set(item.name, item);
                    group.options = [...options.values()];
                } catch {
                    // Keep the original selection available when loading peers fails.
                }
            }),
        );
    }
};

const refreshGroup = async (group: UpgradeGroup, version = generation) => {
    try {
        const res = await searchTasks(
            { page: 1, pageSize: 1, type: 'Image', status: '', taskID: group.taskID },
            operateNode,
        );
        if (!visible.value || version !== generation) return false;
        const task = res.data.items?.[0];
        if (!task) return false;
        group.state = 'submitted';
        group.taskStatus = task.status;
        group.currentStep = task.currentStep;
        return true;
    } catch {
        // A failed status request does not imply that submitting the upgrade failed.
        return false;
    }
};
const needsRefresh = (group: UpgradeGroup) =>
    group.state === 'unknown' ||
    (group.state === 'submitted' && !['Success', 'Failed', 'Canceled'].includes(group.taskStatus));
const poll = async (version: number) => {
    if (!visible.value || version !== generation) return;
    const pending = form.groups.filter(needsRefresh);
    // Bound request concurrency without adding another API or polling abstraction.
    for (let index = 0; index < pending.length; index += 3) {
        await Promise.all(pending.slice(index, index + 3).map((group) => refreshGroup(group, version)));
        if (!visible.value || version !== generation) return;
    }
    if (form.groups.some(needsRefresh)) {
        timer = setTimeout(() => poll(version), 3000);
    }
};
const submit = async () => {
    if (loading.value || !(await formRef.value?.validate().catch(() => false))) return;
    if (loading.value || !canSubmit.value) return;
    loading.value = true;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('container.upgradeWarning2'),
            i18n.global.t(isBatch.value ? 'container.batchUpgrade' : 'commons.button.upgrade'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        );
    } catch {
        loading.value = false;
        return;
    }
    started.value = true;
    clearTimeout(timer);
    const version = ++generation;
    const forcePull = form.forcePull;
    const pending = form.groups.filter((group) => group.names.length > 0 && isGroupEditable(group));
    try {
        for (const group of pending) {
            group.state = 'submitting';
            try {
                await upgradeContainer(
                    {
                        taskID: group.taskID,
                        names: [...group.names],
                        image: group.targetImage.trim(),
                        forcePull,
                    },
                    operateNode,
                );
                group.state = 'submitted';
            } catch (error) {
                if (!(await refreshGroup(group))) {
                    const failure = error as { code?: number; response?: { status?: number } };
                    const rejected =
                        typeof failure.code === 'number' ||
                        ((failure.response?.status || 0) >= 400 && (failure.response?.status || 0) < 500);
                    group.state = rejected ? 'failed' : 'unknown';
                }
            }
        }
    } finally {
        loading.value = false;
        bus.emit('refreshTask', true);
        if (!isBatch.value && form.groups[0]?.state === 'submitted') {
            openLog(form.groups[0]);
            visible.value = false;
        } else {
            void poll(version);
        }
    }
};
const statusLabel = (group: UpgradeGroup) => {
    const t = i18n.global.t;
    if (group.state === 'submitting') return t('container.batchUpgradeSubmitting');
    if (group.state === 'failed') return t('container.batchUpgradeSubmitFailed');
    if (group.state === 'unknown') return t('container.batchUpgradeUnknown');
    if (group.taskStatus === 'Success') return t('commons.status.success');
    if (group.taskStatus === 'Failed') return t('commons.status.failed');
    if (group.taskStatus === 'Canceled') return t('container.batchUpgradeCanceled');
    if (group.currentStep === 'ContainerUpgradeWaiting') return t('container.batchUpgradeWaiting');
    if (group.taskStatus === 'Executing') return t('container.batchUpgradeRunning');
    return t('container.batchUpgradeSubmitted');
};
const statusType = (group: UpgradeGroup) => {
    if (group.state === 'failed' || group.taskStatus === 'Failed') return 'danger';
    if (group.taskStatus === 'Success') return 'success';
    if (group.state === 'unknown') return 'warning';
    return 'info';
};
const openLog = (group: UpgradeGroup) => taskLogRef.value?.openWithTaskID(group.taskID, true, operateNode);
const beforeClose = (done: () => void) => {
    if (!loading.value) done();
};
const handleClose = () => {
    clearTimeout(timer);
    generation++;
    emit('search');
};
onBeforeUnmount(() => {
    clearTimeout(timer);
    generation++;
});
defineExpose({ acceptParams });
</script>

<style scoped>
.upgrade-group + .upgrade-group {
    border-top: 1px solid var(--el-border-color);
    padding-top: 20px;
}
.group-name {
    display: inline-block;
    font-weight: 600;
    color: var(--el-text-color-primary);
    font-size: var(--el-font-size-base);
    overflow-wrap: anywhere;
}
:global(html.dark) .group-name {
    color: #fff;
}
.app-warning {
    --el-alert-title-font-size: var(--el-font-size-base);
    --el-alert-title-with-description-font-size: var(--el-font-size-base);
    --el-alert-description-font-size: var(--el-font-size-base);
}
.image-name {
    max-width: 100%;
    height: auto;
    min-height: 24px;
    font-size: var(--el-font-size-base);
    line-height: 22px;
    white-space: normal;
    overflow-wrap: anywhere;
}
</style>
