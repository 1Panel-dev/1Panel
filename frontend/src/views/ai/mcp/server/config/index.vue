<template>
    <el-drawer
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        v-model="open"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('menu.config')" :back="handleClose" />
        </template>
        <el-input type="textarea" :autosize="{ minRows: 10, maxRows: 20 }" v-model="prettyJson" readonly />
        <CopyButton :content="prettyJson" class="mt-2" />
        <template #footer>
            <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { AI } from '@/api/interface/ai';
import { ref } from 'vue';

const open = ref(false);
const jsonObj = ref({
    mcpServers: {},
});
const prettyJson = ref('');
const handleClose = () => {
    open.value = false;
};

const acceptParams = (mcpServer: AI.McpServer) => {
    jsonObj.value.mcpServers = {};
    jsonObj.value.mcpServers[mcpServer.name] = {
        url: mcpServer.baseUrl + mcpServer.ssePath,
    };
    if (mcpServer.environments) {
        jsonObj.value.mcpServers[mcpServer.name].env = {};
        for (const env of mcpServer.environments) {
            jsonObj.value.mcpServers[mcpServer.name].env[env.key] = env.value;
        }
    }
    prettyJson.value = JSON.stringify(jsonObj.value, null, 4);
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
