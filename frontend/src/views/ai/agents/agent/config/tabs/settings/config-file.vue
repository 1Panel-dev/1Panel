<template>
    <div v-loading="loading">
        <CodemirrorPro
            v-model="form.content"
            mode="json"
            :lineWrapping="true"
            :heightDiff="360"
            :placeholder="t('commons.msg.noneData')"
        />
        <div class="mt-4">
            <el-button type="primary" :loading="saving" @click="confirmSave">
                {{ t('commons.button.save') }}
            </el-button>
        </div>
        <ConfirmDialog ref="confirmDialogRef" @confirm="saveConfig" />
    </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { getAgentConfigFile, updateAgentConfigFile } from '@/api/modules/ai';
import { MsgError, MsgSuccess } from '@/utils/message';

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const agentId = ref(0);
const confirmDialogRef = ref();

const form = reactive({
    content: '',
});

const load = async (id: number) => {
    agentId.value = id;
    loading.value = true;
    try {
        const res = await getAgentConfigFile({ agentId: id });
        form.content = res.data?.content || '';
    } finally {
        loading.value = false;
    }
};

const confirmSave = async () => {
    if (!agentId.value) {
        return;
    }
    try {
        JSON.parse(form.content);
    } catch {
        MsgError(t('commons.rule.formatErr'));
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
