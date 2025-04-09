<template>
    <el-drawer
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        v-model="open"
        size="50%"
    >
        <template #header>
            <DrawerHeader
                :header="$t('mcp.' + mode)"
                :hideResource="mode == 'create'"
                :resource="mcpServer.name"
                :back="handleClose"
            />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="mcpServerForm" label-position="top" :model="mcpServer" label-width="125px" :rules="rules">
                    <el-form-item>
                        <el-button @click="importRef.acceptParams()" type="primary" plain>
                            {{ $t('mcp.importMcpJson') }}
                        </el-button>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model="mcpServer.name" :disabled="mode == 'edit'" />
                    </el-form-item>
                    <el-form-item :label="$t('runtime.runScript')" prop="command">
                        <el-input
                            v-model="mcpServer.command"
                            type="textarea"
                            :rows="3"
                            :placeholder="$t('mcp.commandPlaceHolder')"
                        ></el-input>
                        <span class="input-help">
                            {{ $t('mcp.commandHelper', ['@modelcontextprotocol/server-github']) }}
                        </span>
                    </el-form-item>
                    <div>
                        <el-text>{{ $t('mcp.environment') }}</el-text>
                        <div class="mt-1">
                            <el-row :gutter="20" v-for="(env, index) in mcpServer.environments" :key="index">
                                <el-col :span="8">
                                    <el-form-item :prop="`environments.${index}.key`" :rules="rules.key">
                                        <el-input v-model="env.key" :placeholder="$t('mcp.envKey')" />
                                    </el-form-item>
                                </el-col>
                                <el-col :span="8">
                                    <el-form-item :prop="`environments.${index}.value`" :rules="rules.value">
                                        <el-input v-model="env.value" :placeholder="$t('mcp.envValue')" />
                                    </el-form-item>
                                </el-col>
                                <el-col :span="4">
                                    <el-form-item>
                                        <el-button type="primary" @click="removeEnv(index)" link class="mt-1">
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                            <el-row :gutter="20">
                                <el-col :span="4">
                                    <el-button class="mb-2" @click="addEnv">{{ $t('commons.button.add') }}</el-button>
                                </el-col>
                            </el-row>
                        </div>
                    </div>
                    <Volumes :volumes="mcpServer.volumes" class="mb-2" />
                    <el-row :gutter="20">
                        <el-col :span="8">
                            <el-form-item :label="$t('commons.table.port')" prop="port">
                                <el-input v-model.number="mcpServer.port" />
                            </el-form-item>
                        </el-col>
                        <el-col :span="6">
                            <el-form-item :label="$t('app.allowPort')" prop="hostIP">
                                <el-switch
                                    v-model="mcpServer.hostIP"
                                    :active-value="'0.0.0.0'"
                                    :inactive-value="'127.0.0.1'"
                                />
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-form-item :label="$t('mcp.baseUrl')" prop="url">
                        <el-input v-model.trim="mcpServer.url">
                            <template #prepend>
                                <el-select v-model="mcpServer.protocol" class="pre-select">
                                    <el-option label="http" value="http://" />
                                    <el-option label="https" value="https://" />
                                </el-select>
                            </template>
                        </el-input>
                        <span class="input-help">
                            {{ $t('mcp.baseUrlHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('app.containerName')" prop="containerName">
                        <el-input v-model.trim="mcpServer.containerName"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('mcp.ssePath')" prop="ssePath">
                        <el-input v-model.trim="mcpServer.ssePath"></el-input>
                        <span class="input-help">
                            {{ $t('mcp.ssePathHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(mcpServerForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
    <Import ref="importRef" @confirm="getImport" />
</template>

<script lang="ts" setup>
import { AI } from '@/api/interface/ai';
import { createMcpServer, getMcpDomain, updateMcpServer } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { ref, watch } from 'vue';
import Volumes from '../volume/index.vue';
import Import from '../import/index.vue';

const open = ref(false);
const mode = ref('create');
const loading = ref(false);
const mcpServerForm = ref();
const importRef = ref();
const newMcpServer = () => {
    return {
        id: 0,
        name: '',
        port: 8000,
        status: '',
        message: '',
        baseUrl: '',
        ssePath: '',
        command: '',
        containerName: '',
        environments: [],
        volumes: [],
        hostIP: '127.0.0.1',
        protocol: 'http://',
        url: '',
    };
};
const em = defineEmits(['close']);
const mcpServer = ref(newMcpServer());
const rules = ref({
    name: [Rules.requiredInput, Rules.appName],
    command: [Rules.requiredInput],
    port: [Rules.requiredInput, Rules.port],
    containerName: [Rules.requiredInput],
    url: [Rules.requiredInput],
    ssePath: [Rules.requiredInput],
    key: [Rules.requiredInput],
    value: [Rules.requiredInput],
});
const hasWebsite = ref(false);

const acceptParams = async (params: AI.McpServer) => {
    hasWebsite.value = false;
    mode.value = params.id ? 'edit' : 'create';
    let mcpDomainRes;
    try {
        mcpDomainRes = await getMcpDomain();
        if (mcpDomainRes.data.connUrl != '') {
            hasWebsite.value = true;
        }
    } catch (error) {}

    if (mode.value == 'edit') {
        mcpServer.value = params;
        if (!mcpServer.value.environments) {
            mcpServer.value.environments = [];
        }
        if (!mcpServer.value.volumes) {
            mcpServer.value.volumes = [];
        }
        const parts = mcpServer.value.baseUrl.split(/(https?:\/\/)/).filter(Boolean);
        mcpServer.value.protocol = parts[0];
        mcpServer.value.url = parts[1];
    } else {
        mcpServer.value = newMcpServer();
        if (params.port) {
            mcpServer.value.port = params.port;
        }
        if (mcpDomainRes.data && mcpDomainRes.data.connUrl != '') {
            const parts = mcpDomainRes.data.connUrl.split(/(https?:\/\/)/).filter(Boolean);
            mcpServer.value.protocol = parts[0];
            mcpServer.value.url = parts[1];
            mcpServer.value.baseUrl = mcpDomainRes.data.connUrl;
        }
    }
    open.value = true;
};

watch(
    () => mcpServer.value.name,
    (newVal) => {
        if (newVal && mode.value == 'create') {
            mcpServer.value.containerName = newVal;
            mcpServer.value.ssePath = '/' + newVal;
        }
    },
    { deep: true },
);

const addEnv = () => {
    mcpServer.value.environments.push({
        key: '',
        value: '',
    });
};

const removeEnv = (index: number) => {
    mcpServer.value.environments.splice(index, 1);
};

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const getImport = async (data: AI.ImportMcpServer[]) => {
    if (!data) {
        return;
    }
    const importServer = data[0];
    mcpServer.value.name = importServer.name;
    mcpServer.value.containerName = importServer.containerName;
    mcpServer.value.ssePath = importServer.ssePath;
    mcpServer.value.command = importServer.command;
    mcpServer.value.environments = importServer.environments || [];
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) {
            return;
        }
        let request = true;
        if (mcpServer.value.hostIP != '0.0.0.0' && !hasWebsite.value) {
            await ElMessageBox.confirm(i18n.global.t('app.installWarn'), i18n.global.t('app.checkTitle'), {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            }).catch(() => {
                request = false;
            });
        }
        if (!request) {
            return;
        }
        try {
            loading.value = true;
            mcpServer.value.baseUrl = mcpServer.value.protocol + mcpServer.value.url;
            if (mode.value == 'create') {
                await createMcpServer(mcpServer.value);
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            } else {
                await updateMcpServer(mcpServer.value);
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            }
            handleClose();
        } finally {
            loading.value = false;
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
