<template>
    <DialogPro
        v-model="visible"
        class="firewall-rule-dialog firewall-initialize-dialog"
        :title="
            starting
                ? $t('firewall.firewallHelper', [$t('commons.button.start')])
                : initialize
                  ? $t('commons.button.init')
                  : $t('firewall.importRuleBackup')
        "
        :show-close="!loading"
        :close-on-press-escape="!loading"
    >
        <template #content>
            <div v-loading="loading" class="firewall-rule-dialog__content">
                <slot name="before-options" />
                <p v-if="message && starting" class="firewall-initialize-dialog__description">{{ message }}</p>
                <el-alert v-else-if="message" :title="message" type="warning" :closable="false" show-icon />
                <el-alert
                    v-if="backupLoadFailed"
                    :title="$t('firewall.ruleBackupLoadFailed')"
                    type="warning"
                    :closable="false"
                    show-icon
                >
                    <el-button type="primary" link :disabled="loading" @click="loadBackups">
                        {{ $t('commons.button.retry') }}
                    </el-button>
                </el-alert>
                <el-radio-group
                    v-if="initialize"
                    v-model="mode"
                    class="firewall-rule-dialog__modes"
                    :disabled="loading"
                >
                    <el-radio value="initialize" border>
                        <span class="firewall-rule-dialog__choice-title">{{ initializeLabel }}</span>
                    </el-radio>
                    <el-radio value="restore" border>
                        <span class="firewall-rule-dialog__choice-title">{{ restoreLabel }}</span>
                    </el-radio>
                </el-radio-group>
                <div v-if="loaded && mode === 'restore'" class="firewall-rule-dialog__directory">
                    <el-icon><FolderOpened /></el-icon>
                    <span>{{ $t('firewall.backupDirectoryHelper', [backups.directory]) }}</span>
                </div>
                <div v-if="mode === 'restore'" class="firewall-initialize-dialog__backups">
                    <el-radio-group
                        v-if="backups.files.length"
                        v-model="selected"
                        class="firewall-rule-dialog__files"
                        :disabled="loading"
                        :aria-label="$t('firewall.importRuleBackup')"
                    >
                        <el-radio v-for="file in backups.files" :key="file.name" :value="file.name" border>
                            <span class="firewall-rule-dialog__file-name">
                                <el-icon><Document /></el-icon>
                                <span class="min-w-0">{{ file.name }}</span>
                                <el-button
                                    v-permission
                                    class="shrink-0"
                                    type="primary"
                                    link
                                    size="small"
                                    :disabled="loading"
                                    :aria-label="`${$t('firewall.downloadRuleBackup')}: ${file.name}`"
                                    @click.stop.prevent="
                                        downloadFile(`${backups.directory}/${file.name}`, globalStore.currentNode)
                                    "
                                >
                                    {{ $t('commons.button.download') }}
                                </el-button>
                            </span>
                            <span class="firewall-rule-dialog__file-meta">
                                <el-tag size="small" type="info" effect="plain">{{ file.provider }}</el-tag>
                                <span>{{ $t('firewall.backupRuleCount') }}: {{ file.ruleCount }}</span>
                                <span>{{ new Date(file.modifiedAt).toLocaleString() }}</span>
                            </span>
                        </el-radio>
                    </el-radio-group>
                    <el-empty
                        v-else-if="!initialize && loaded && !backups.files.length"
                        :description="$t('firewall.noRuleBackup')"
                        :image-size="64"
                    />
                </div>
            </div>
        </template>
        <template #footer>
            <slot name="confirmation" />
            <el-button :disabled="loading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                type="primary"
                :loading="loading"
                :disabled="loading || (mode === 'restore' && (!loaded || !selected))"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
    <TaskLog ref="taskLogRef" @close="emit('complete')" />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import i18n from '@/lang';
import { Document, FolderOpened } from '@element-plus/icons-vue';
import DialogPro from '@/components/dialog-pro/index.vue';
import { Firewall } from '@/api/interface/firewall';
import {
    createFirewallRules,
    listFirewallRuleBackups,
    enableForwarding,
    operateDockerPortGuard,
} from '@/api/modules/firewall';
import TaskLog from '@/components/log/task/index.vue';
import { GlobalStore } from '@/store';
import { downloadFile } from '@/utils/file';

const globalStore = GlobalStore();
const emit = defineEmits<{
    (event: 'initialize'): void;
    (event: 'restore', backupFile: string): void;
    (event: 'complete'): void;
}>();
const visible = ref(false);
const loading = ref(false);
const loaded = ref(false);
const backupLoadFailed = ref(false);
const message = ref('');
const initialize = ref(true);
const starting = ref(false);
const initializeLabel = computed(() =>
    i18n.global.t(starting.value ? 'firewall.startOnly' : 'firewall.initializeOnly'),
);
const restoreLabel = computed(() =>
    i18n.global.t(starting.value ? 'firewall.startAndRestore' : 'firewall.restoreAllRules'),
);
const subsystem = ref<Firewall.BackendSubsystem>('system');
const mode = ref<'restore' | 'initialize'>('initialize');
const selected = ref('');
const backups = ref<Firewall.RuleBackups>({ directory: '', files: [] });
const taskLogRef = ref<InstanceType<typeof TaskLog>>();

const acceptParams = async (
    warning: string,
    withInitialization = true,
    target: Firewall.BackendSubsystem = 'system',
    start = false,
) => {
    subsystem.value = target;
    starting.value = start;
    initialize.value = withInitialization;
    message.value = warning;
    visible.value = true;
    await loadBackups();
};

const loadBackups = async () => {
    if (loading.value) return;
    backups.value = { directory: '', files: [] };
    selected.value = '';
    mode.value = initialize.value ? 'initialize' : 'restore';
    loaded.value = false;
    backupLoadFailed.value = false;
    loading.value = true;
    try {
        backups.value = (await listFirewallRuleBackups(subsystem.value)).data;
        selected.value = backups.value.files[0]?.name || '';
        mode.value = !starting.value && (selected.value || !initialize.value) ? 'restore' : 'initialize';
        loaded.value = true;
    } catch {
        backupLoadFailed.value = true;
    } finally {
        loading.value = false;
    }
};

const submit = async () => {
    if (loading.value) return;
    if (mode.value === 'initialize') {
        visible.value = false;
        emit('initialize');
        return;
    }
    if (!loaded.value || !selected.value) return;
    if (starting.value) {
        visible.value = false;
        emit('restore', selected.value);
        return;
    }
    loading.value = true;
    try {
        const response =
            subsystem.value === 'forwarding'
                ? await enableForwarding(undefined, selected.value)
                : subsystem.value === 'docker'
                  ? await operateDockerPortGuard('initialize', undefined, selected.value)
                  : await createFirewallRules({ initialize: initialize.value, backupFile: selected.value, items: [] });
        const result = response.data;
        visible.value = false;
        if (result.queued && result.taskID) {
            taskLogRef.value?.openWithTaskID(result.taskID, true);
        } else {
            emit('complete');
        }
    } finally {
        loading.value = false;
    }
};

defineExpose({ acceptParams });
</script>

<style lang="scss" src="./rule-dialog.scss"></style>
