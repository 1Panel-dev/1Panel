<template>
    <el-form label-position="top">
        <el-form-item>
            <el-button type="primary" :loading="loggingIn" @click="loginChannel">
                {{ t('aiTools.agents.scanConnect') }}
            </el-button>
        </el-form-item>
        <el-alert type="info" :closable="false" :title="t('aiTools.agents.scanConnectHelper')" />
    </el-form>
    <TaskLog ref="loginTaskLogRef" />
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { loginAgentWeixinChannel } from '@/api/modules/ai';
import TaskLog from '@/components/log/task/index.vue';
import { newUUID } from '@/utils/id';

const props = defineProps<{
    agentId: number;
}>();
const { t } = useI18n();
const loggingIn = ref(false);
const loginTaskLogRef = ref();

const load = async () => {};

const loginChannel = async () => {
    const taskID = newUUID();
    loggingIn.value = true;
    try {
        await loginAgentWeixinChannel({
            agentId: props.agentId,
            taskID,
        });
        loginTaskLogRef.value?.openWithTaskID(taskID);
    } finally {
        loggingIn.value = false;
    }
};

defineExpose({
    load,
});
</script>
