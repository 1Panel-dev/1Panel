<template>
    <DrawerPro
        v-model="open"
        @close="handleClose"
        size="large"
        :header="$t('commons.button.' + item.operate) + $t('website.loadBalance')"
        :resource="item.operate == 'create' ? '' : item.name"
    >
        <el-form ref="lbForm" label-position="top" :model="item" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="item.name" :disabled="item.operate === 'edit'"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.algorithm')" prop="algorithm">
                <el-select v-model="item.algorithm">
                    <el-option
                        v-for="(algorithm, index) in Algorithms"
                        :label="algorithm.label"
                        :key="index"
                        :value="algorithm.value"
                    ></el-option>
                </el-select>
                <span class="input-help">{{ getHelper(item.algorithm) }}</span>
            </el-form-item>

            <div>
                <el-card v-for="(server, index) of item.servers" :key="index" class="server-card" shadow="hover">
                    <template #header>
                        <div class="card-header">
                            <span class="server-title">
                                <el-icon><Monitor /></el-icon>
                                <span class="ml-2">{{ $t('website.server') }} - {{ server.server }}</span>
                            </span>
                            <el-button
                                v-if="item.servers.length > 1"
                                text
                                type="danger"
                                icon="Delete"
                                size="small"
                                @click="removeServer(index)"
                            ></el-button>
                        </div>
                    </template>

                    <div>
                        <el-row :gutter="16">
                            <el-col :span="24">
                                <el-form-item
                                    :label="$t('setting.address')"
                                    :prop="`servers.${index}.server`"
                                    :rules="rules.server"
                                >
                                    <el-input
                                        v-model="item.servers[index].server"
                                        :placeholder="'example.com:8080'"
                                    ></el-input>
                                </el-form-item>
                            </el-col>
                        </el-row>
                        <el-row :gutter="16">
                            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                                <el-form-item
                                    :label="$t('website.weight')"
                                    :prop="`servers.${index}.weight`"
                                    :rules="rules.weight"
                                >
                                    <el-input type="number" v-model.number="item.servers[index].weight"></el-input>
                                </el-form-item>
                            </el-col>
                            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                                <el-form-item
                                    :label="$t('website.strategy')"
                                    :prop="`servers.${index}.flag`"
                                    :rules="rules.flag"
                                >
                                    <el-select v-model="item.servers[index].flag" clearable>
                                        <el-option
                                            v-for="flag in getStatusStrategy()"
                                            :label="flag.label"
                                            :key="flag.value"
                                            :value="flag.value"
                                        ></el-option>
                                    </el-select>
                                </el-form-item>
                            </el-col>
                        </el-row>
                        <el-row :gutter="16">
                            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                                <el-form-item
                                    :label="$t('website.maxFails')"
                                    :prop="`servers.${index}.maxFails`"
                                    :rules="rules.maxFails"
                                >
                                    <el-input type="number" v-model.number="item.servers[index].maxFails">
                                        <template #append>{{ $t('commons.units.time') }}</template>
                                    </el-input>
                                </el-form-item>
                            </el-col>
                            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                                <el-form-item :prop="`servers.${index}.failTimeout`" :rules="rules.failTimeout">
                                    <template #label>
                                        <span class="inline-flex items-center">
                                            {{ $t('website.failTimeout') }}
                                            <el-tooltip :content="$t('website.failTimeoutHelper')" placement="top">
                                                <el-icon><QuestionFilled /></el-icon>
                                            </el-tooltip>
                                        </span>
                                    </template>
                                    <el-input type="number" v-model.number="item.servers[index].failTimeout">
                                        <template #append>
                                            <el-select
                                                v-model.number="item.servers[index].failTimeoutUnit"
                                                class="!w-24"
                                            >
                                                <el-option
                                                    v-for="(unit, index) in Units"
                                                    :key="index"
                                                    :label="unit.label"
                                                    :value="unit.value"
                                                />
                                            </el-select>
                                        </template>
                                    </el-input>
                                </el-form-item>
                            </el-col>
                        </el-row>
                        <el-row :gutter="16">
                            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                                <el-form-item
                                    :label="$t('website.maxConns')"
                                    :prop="`servers.${index}.maxConns`"
                                    :rules="rules.maxConns"
                                >
                                    <el-input type="number" v-model.number="item.servers[index].maxConns"></el-input>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </div>
                </el-card>

                <el-button class="add-server-btn" type="primary" plain @click="addServer" icon="Plus">
                    {{ $t('commons.button.add') + $t('website.server') }}
                </el-button>
            </div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(lbForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { createLoadBalance, updateLoadBalance } from '@/api/modules/website';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { getAlgorithms, getStatusStrategy } from '@/global/mimetype';
import { Website } from '@/api/interface/website';
import { Monitor, QuestionFilled } from '@element-plus/icons-vue';
import { Units } from '@/global/mimetype';

const rules = ref<any>({
    name: [Rules.appName],
    algorithm: [Rules.requiredSelect],
    server: [Rules.requiredInput],
    weight: [checkNumberRange(0, 100)],
    servers: {
        type: Array,
    },
    maxFails: [checkNumberRange(1, 1000)],
    maxConns: [checkNumberRange(0, 10000)],
    failTimeout: [checkNumberRange(1, 300)],
});

interface LoadBalanceOperate {
    websiteID: number;
    operate: string;
    upstream?: Website.NginxUpstream;
}

const lbForm = ref<FormInstance>();

const initServer = () => ({
    server: '',
    weight: undefined,
    maxFails: undefined,
    maxConns: undefined,
    failTimeout: undefined,
    failTimeoutUnit: 's',
    flag: '',
});

const open = ref(false);
const loading = ref(false);
const item = ref({
    websiteID: 0,
    name: '',
    operate: 'create',
    servers: [],
    algorithm: 'default',
    flag: '',
});

const em = defineEmits(['close']);
const handleClose = () => {
    lbForm.value?.resetFields();
    open.value = false;
    em('close', false);
};

const helper = ref();
const Algorithms = getAlgorithms();
const getHelper = (key: string) => {
    Algorithms.forEach((algorithm) => {
        if (algorithm.value === key) {
            helper.value = algorithm.placeHolder;
        }
    });
    return helper.value;
};

const addServer = () => {
    item.value.servers.push(initServer());
};

const removeServer = (index: number) => {
    item.value.servers.splice(index, 1);
};

const acceptParams = async (req: LoadBalanceOperate) => {
    item.value.websiteID = req.websiteID;
    item.value.operate = req.operate;
    if (req.operate == 'edit') {
        item.value.operate = 'edit';
        item.value.name = req.upstream?.name || '';
        item.value.algorithm = req.upstream?.algorithm || 'default';
        let servers = [];
        req.upstream?.servers?.forEach((server) => {
            const weight = server.weight == 0 ? undefined : server.weight;
            const maxFails = server.maxFails == 0 ? undefined : server.maxFails;
            const maxConns = server.maxConns == 0 ? undefined : server.maxConns;
            const failTimeout = server.failTimeout == 0 ? undefined : server.failTimeout;
            const failTimeoutUnit = server.failTimeoutUnit || 's';
            servers.push({
                server: server.server,
                weight: weight,
                maxFails: maxFails,
                maxConns: maxConns,
                failTimeout: failTimeout,
                failTimeoutUnit: failTimeoutUnit,
                flag: server.flag,
            });
        });
        item.value.servers = servers;
    } else {
        item.value.name = '';
        item.value.servers = [initServer()];
    }
    open.value = true;
};

const handleServers = () => {
    for (const server of item.value.servers) {
        if (!server.weight || server.weight == '') {
            server.weight = 0;
        }
        if (!server.maxFails || server.maxFails == '') {
            server.maxFails = 0;
        }
        if (!server.maxConns || server.maxConns == '') {
            server.maxConns = 0;
        }
        if (!server.failTimeout || server.failTimeout == '') {
            server.failTimeout = 0;
        }
    }
};

const rollBackServers = () => {
    for (const server of item.value.servers) {
        if (server.weight == 0) {
            server.weight = undefined;
        }
        if (server.maxFails == 0) {
            server.maxFails = undefined;
        }
        if (server.maxConns == 0) {
            server.maxConns = undefined;
        }
        if (server.failTimeout == 0) {
            server.failTimeout = undefined;
        }
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) {
            return;
        }
        let checkBackup = false;
        if (item.value.algorithm == 'ip_hash') {
            checkBackup = true;
        }
        for (const server of item.value.servers) {
            if (checkBackup && server.flag == 'backup') {
                MsgError(i18n.global.t('website.ipHashBackupErr'));
                return;
            }
        }
        handleServers();
        loading.value = true;
        try {
            if (item.value.operate === 'edit') {
                await updateLoadBalance(item.value);
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            } else {
                await createLoadBalance(item.value);
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            }
            handleClose();
        } catch {
            rollBackServers();
        } finally {
            loading.value = false;
        }
    });
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.server-card {
    margin-bottom: 16px;
    border-radius: 8px;
    transition: all 0.3s ease;

    &:hover {
        border-color: var(--el-color-primary);
    }

    :deep(.el-card__header) {
        padding: 12px 20px;
        background-color: var(--el-fill-color-light);
    }
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 20px;

    .server-title {
        display: flex;
        align-items: center;
        font-weight: 500;
        font-size: 14px;
        color: var(--el-text-color-primary);
    }
}

.add-server-btn {
    width: 100%;
    height: 48px;
    border-style: dashed;
    font-size: 14px;
    margin-top: 8px;

    &:hover {
        border-color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
    }
}
</style>
