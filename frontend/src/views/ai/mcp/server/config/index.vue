<template>
    <DrawerPro v-model="open" :header="$t('menu.config')" @close="handleClose" size="normal">
        <codemirror
            :autofocus="true"
            :placeholder="$t('commons.msg.noneData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="height: 300px"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            v-model="prettyJson"
            :disabled="true"
        />
        <CopyButton :content="prettyJson" class="mt-2" />
        <template #footer>
            <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { AI } from '@/api/interface/ai';
import { ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';

const extensions = [javascript(), oneDark];

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
    prettyJson.value = JSON.stringify(jsonObj.value, null, 2);
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
