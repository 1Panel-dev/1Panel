<template>
    <el-form v-loading="deleting" label-position="top">
        <el-form-item v-if="configured">
            <el-button v-permission type="danger" plain :loading="deleting" @click="deleteChannel">
                {{ t('commons.button.delete') }}
            </el-button>
        </el-form-item>
        <el-form-item>
            <el-button v-permission type="primary" :loading="loggingIn" @click="loginChannel">
                {{ t('aiTools.agents.scanConnect') }}
            </el-button>
        </el-form-item>
        <el-alert type="info" :closable="false" :title="t('aiTools.agents.scanConnectHelper')" />
    </el-form>
    <TaskLog ref="loginTaskLogRef" />
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { deleteAgentChannelConfig, getAgentWeixinConfig, loginAgentWeixinChannel } from '@/api/modules/ai';
import TaskLog from '@/components/log/task/index.vue';
import { newUUID } from '@/utils/id';
import { MsgSuccess } from '@/utils/message';

const props = defineProps<{
    agentId: number;
}>();
const { t } = useI18n();
const loggingIn = ref(false);
const deleting = ref(false);
const loginTaskLogRef = ref();
const configured = ref(false);

const load = async () => {
    const res = await getAgentWeixinConfig({ agentId: props.agentId });
    configured.value = !!res.data?.enabled;
};

const loginChannel = async () => {
    const taskID = newUUID();
    loggingIn.value = true;
    try {
        await loginAgentWeixinChannel({
            agentId: props.agentId,
            taskID,
        });
        configured.value = true;
        loginTaskLogRef.value?.openWithTaskID(taskID);
    } finally {
        loggingIn.value = false;
    }
};

const deleteChannel = async () => {
    await ElMessageBox.confirm(
        t('aiTools.agents.channelDeleteConfirm', [t('aiTools.agents.weixin')]),
        t('commons.msg.infoTitle'),
        {
            confirmButtonText: t('commons.button.confirm'),
            cancelButtonText: t('commons.button.cancel'),
            type: 'warning',
        },
    );
    deleting.value = true;
    try {
        await deleteAgentChannelConfig({
            agentId: props.agentId,
            type: 'weixin',
        });
        await load();
        MsgSuccess(t('aiTools.agents.successAndRestart', [t('commons.msg.deleteSuccess')]));
    } finally {
        deleting.value = false;
    }
};

defineExpose({
    load,
});
</script>
