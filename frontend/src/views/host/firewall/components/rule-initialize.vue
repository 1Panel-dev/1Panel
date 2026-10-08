<template>
    <DialogPro
        v-model="visible"
        class="firewall-rule-dialog"
        :title="initialize ? $t('commons.button.init') : $t('firewall.importRuleBackup')"
    >
        <template #content>
            <div v-loading="loading" class="firewall-rule-dialog__content">
                <el-alert v-if="message" :title="message" type="warning" :closable="false" show-icon />
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
                    <el-radio value="restore" border :disabled="!loaded || backups.files.length === 0">
                        {{ $t('firewall.restoreAllRules') }}
                    </el-radio>
                    <el-radio value="initialize" border>{{ $t('firewall.initializeOnly') }}</el-radio>
                </el-radio-group>
                <div v-if="loaded" class="firewall-rule-dialog__directory">
                    <el-icon><FolderOpened /></el-icon>
                    <span>{{ $t('firewall.backupDirectoryHelper', [backups.directory]) }}</span>
                </div>
                <el-radio-group
                    v-if="mode === 'restore' && backups.files.length"
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
                    v-else-if="loaded && !backups.files.length"
                    :description="$t('firewall.noRuleBackup')"
                    :image-size="64"
                />
            </div>
        </template>
        <template #footer>
            <el-button :disabled="loading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                type="primary"
                :loading="loading"
                :disabled="loading || (mode === 'restore' && (!loaded || !selected))"
                @click="submit"
            >
                {{
                    initialize
                        ? mode === 'restore'
                            ? $t('firewall.restoreAllRules')
                            : $t('firewall.initializeOnly')
                        : $t('commons.button.import')
                }}
            </el-button>
        </template>
    </DialogPro>
    <TaskLog ref="taskLogRef" @close="emit('complete')" />
</template>

<script setup lang="ts">
import { ref } from 'vue';
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
const emit = defineEmits<{ (event: 'initialize'): void; (event: 'complete'): void }>();
const visible = ref(false);
const loading = ref(false);
const loaded = ref(false);
const backupLoadFailed = ref(false);
const message = ref('');
const initialize = ref(true);
const subsystem = ref<Firewall.BackendSubsystem>('system');
const mode = ref<'restore' | 'initialize'>('initialize');
const selected = ref('');
const backups = ref<Firewall.RuleBackups>({ directory: '', files: [] });
const taskLogRef = ref<InstanceType<typeof TaskLog>>();

const acceptParams = async (
    warning: string,
    withInitialization = true,
    target: Firewall.BackendSubsystem = 'system',
) => {
    subsystem.value = target;
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
        mode.value = selected.value || !initialize.value ? 'restore' : 'initialize';
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
