<template>
    <el-dialog v-model="submitVisible" :destroy-on-close="true" :close-on-click-modal="false" width="40%">
        <template #header>
            {{ $t('mcp.importMcpJson') }}
        </template>
        <div>
            <el-input
                v-model="mcpServerJson"
                type="textarea"
                :rows="15"
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
            <span class="dialog-footer">
                <el-button @click="onCancel">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onConfirm">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
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
            throw new Error(i18n.global.t('mcp.importMcpJsonError'));
        }
        mcpServerConfig.value = Object.entries(data.mcpServers).map(([name, config]: any) => ({
            name,
            command: [config.command, ...config.args].join(' '),
            environments: data.env ? Object.entries(data.env).map(([key, value]) => ({ key, value })) : [],
            ssePath: '/' + name,
            containerName: name,
        }));
    } catch (error) {
        MsgError(error);
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
