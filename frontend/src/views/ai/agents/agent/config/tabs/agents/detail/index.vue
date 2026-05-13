<template>
    <DialogPro v-model="open" :title="header" size="w-60" @close="handleClose">
        <div v-loading="loading" class="agent-file-drawer">
            <el-alert v-if="!agentId || !workspace" :title="$t('commons.msg.noneData')" type="info" :closable="false" />
            <template v-else>
                <el-tabs v-model="activeFile">
                    <el-tab-pane v-for="item in fileItems" :key="item.name" :name="item.name">
                        <template #label>{{ item.name }}</template>
                        <div class="agent-file-status" v-if="item.error">{{ item.error }}</div>
                        <CodemirrorPro
                            v-model="item.content"
                            :height-diff="470"
                            :line-wrapping="true"
                            :placeholder="item.name"
                        />
                        <ul v-if="getFileDescriptionItems(item.name).length" class="agent-file-description">
                            <li v-for="description in getFileDescriptionItems(item.name)" :key="description">
                                {{ description }}
                            </li>
                        </ul>
                    </el-tab-pane>
                </el-tabs>
            </template>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="saving" @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :loading="loading" :disabled="saving || !agentId || !workspace" @click="reloadFiles">
                    {{ $t('commons.button.refresh') }}
                </el-button>
                <el-button v-permission type="primary" :loading="saving" :disabled="!currentFile" @click="saveAllFiles">
                    {{ $t('aiTools.agents.saveAllMd') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { getAgentRoleMarkdownFiles, updateAgentRoleMarkdownFile } from '@/api/modules/ai';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

interface DialogParams {
    name: string;
    agentId: number;
    workspace: string;
}

interface AgentFileItem {
    name: string;
    content: string;
    error: string;
}

const FILE_NAMES = [
    'AGENTS.md',
    'SOUL.md',
    'USER.md',
    'IDENTITY.md',
    'TOOLS.md',
    'HEARTBEAT.md',
    'BOOT.md',
    'BOOTSTRAP.md',
];
const open = ref(false);
const loading = ref(false);
const saving = ref(false);
const roleName = ref('');
const agentId = ref(0);
const workspace = ref('');
const activeFile = ref(FILE_NAMES[0]);
const fileItems = ref<AgentFileItem[]>([]);

const header = computed(() => `${roleName.value || '-'}`);
const currentFile = computed(() => fileItems.value.find((item) => item.name === activeFile.value));
const getFileDescriptionItems = (name: string): string[] => {
    const descriptions = i18n.global.tm('aiTools.agents.roleMarkdownDescriptions') as Record<string, unknown>;
    const items = descriptions?.[name];
    return Array.isArray(items) ? items.map((item) => String(item)) : [];
};

const buildFiles = (items: AI.AgentRoleMarkdownFileItem[]) =>
    FILE_NAMES.map((name) => ({
        name,
        content: items.find((item) => item.name === name)?.content || '',
        error: '',
    }));

const reloadFiles = async () => {
    if (!agentId.value || !workspace.value) {
        return;
    }
    loading.value = true;
    try {
        const res = await getAgentRoleMarkdownFiles({
            agentId: agentId.value,
            workspace: workspace.value,
        });
        fileItems.value = buildFiles(res.data || []);
        activeFile.value = fileItems.value[0]?.name || FILE_NAMES[0];
    } finally {
        loading.value = false;
    }
};

const acceptParams = async (params: DialogParams) => {
    roleName.value = params.name;
    agentId.value = params.agentId;
    workspace.value = params.workspace;
    open.value = true;
    await reloadFiles();
};

const saveAllFiles = async () => {
    if (!fileItems.value.length) {
        return;
    }
    let restart = false;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('aiTools.agents.roleMarkdownRestartHelper'),
            i18n.global.t('aiTools.agents.saveAllMd'),
            {
                confirmButtonText: i18n.global.t('setting.restartNow'),
                cancelButtonText: i18n.global.t('setting.restartLater'),
                type: 'warning',
                distinguishCancelAndClose: true,
            },
        );
        restart = true;
    } catch (error) {
        if (error === 'cancel') {
            restart = false;
        } else {
            return;
        }
    }
    saving.value = true;
    try {
        await updateAgentRoleMarkdownFile({
            agentId: agentId.value,
            workspace: workspace.value,
            restart,
            files: fileItems.value.map((item) => ({
                name: item.name,
                content: item.content,
            })),
        });
        fileItems.value.forEach((item) => {
            item.error = '';
        });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        saving.value = false;
    }
};

const handleClose = () => {
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.agent-file-drawer {
    min-height: 320px;
}

.agent-file-status {
    margin-bottom: 12px;
    color: var(--el-color-danger);
}

.agent-file-description {
    margin-top: 12px;
    margin-bottom: 0;
    padding-left: 18px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1.6;
}

.agent-file-description li + li {
    margin-top: 4px;
}

.dialog-footer {
    display: flex;
    align-items: center;
    gap: 8px;
}
</style>
