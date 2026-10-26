<template>
    <DialogPro v-model="submitVisible">
        <template #header>
            {{ $t('aiTools.mcp.importMcpJson') }}
        </template>
        <div class="w-full min-w-0">
            <el-input
                v-model="mcpServerJson"
                type="textarea"
                :rows="15"
                class="[&_textarea]:!font-mono [&_textarea]:!text-xs [&_textarea]:!leading-5"
                placeholder='{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-postgres",
        "postgresql://localhost/mydb"
      ]
    }
  }
}'
            ></el-input>
        </div>
        <template #footer>
            <div class="flex flex-wrap justify-end gap-2 [&>.el-button]:!ml-0">
                <el-button @click="onCancel">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button v-permission type="primary" @click="onConfirm">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </div>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { ref } from 'vue';

const submitVisible = ref(false);
const mcpServerJson = ref();
const mcpServerConfig = ref();

const acceptParams = (): void => {
    mcpServerJson.value = '';
    submitVisible.value = true;
};
const emit = defineEmits(['confirm', 'cancel']);

const onConfirm = async () => {
    try {
        const data = JSON.parse(mcpServerJson.value);
        if (!data.mcpServers || typeof data.mcpServers !== 'object') {
            throw new Error(i18n.global.t('aiTools.mcp.importMcpJsonError'));
        }
        mcpServerConfig.value = Object.entries(data.mcpServers).map(([name, config]: any) => ({
            name,
            command: [config.command, ...config.args].join(' '),
            environments: config.env ? Object.entries(config.env).map(([key, value]) => ({ key, value })) : [],
            ssePath: '/' + name,
            containerName: name,
        }));
    } catch (error) {
        return;
    }
    emit('confirm', mcpServerConfig.value);
    submitVisible.value = false;
};

const onCancel = async () => {
    emit('cancel');
    submitVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
