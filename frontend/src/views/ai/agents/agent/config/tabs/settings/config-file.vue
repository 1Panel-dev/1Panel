<template>
    <div v-loading="loading">
        <CodemirrorPro
            :key="editorMode"
            v-model="form.content"
            :mode="editorMode"
            :lineWrapping="true"
            :heightDiff="360"
            :placeholder="t('commons.msg.noneData')"
        />
        <div class="mt-4">
            <el-button v-permission type="primary" :loading="saving" @click="confirmSave">
                {{ t('commons.button.save') }}
            </el-button>
        </div>
        <ConfirmDialog ref="confirmDialogRef" @confirm="saveConfig" />
    </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { AI } from '@/api/interface/ai';
import { getAgentConfigFile, updateAgentConfigFile } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const agentId = ref(0);
const agentType = ref<AI.AgentType>('openclaw');
const confirmDialogRef = ref();

const form = reactive({
    content: '',
});

const editorMode = computed(() => (agentType.value === 'hermes-agent' ? 'yaml' : 'json'));

const load = async (params: { agentId: number; agentType: AI.AgentType }) => {
    agentId.value = params.agentId;
    agentType.value = params.agentType;
    loading.value = true;
    try {
        const res = await getAgentConfigFile({ agentId: params.agentId });
        form.content = res.data?.content || '';
    } finally {
        loading.value = false;
    }
};

const confirmSave = async () => {
    if (!agentId.value) {
        return;
    }
    confirmDialogRef.value?.acceptParams({
        header: t('database.confChange'),
        operationInfo: t('aiTools.agents.configFileRestartHelper'),
        submitInputInfo: t('database.restartNow'),
    });
};

const saveConfig = async () => {
    if (!agentId.value) {
        return;
    }
    saving.value = true;
    try {
        await updateAgentConfigFile({
            agentId: agentId.value,
            content: form.content,
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
    } finally {
        saving.value = false;
    }
};

defineExpose({
    load,
});
</script>
