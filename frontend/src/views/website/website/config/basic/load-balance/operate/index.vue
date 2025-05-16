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
            <el-row :gutter="20" v-for="(server, index) of item.servers" :key="index">
                <el-col :span="7">
                    <el-form-item
                        :label="index == 0 ? $t('setting.address') : ''"
                        :prop="`servers.${index}.server`"
                        :rules="rules.server"
                    >
                        <el-input
                            type="string"
                            v-model="item.servers[index].server"
                            :placeholder="index > 0 ? $t('setting.address') : ''"
                        ></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="3">
                    <el-form-item
                        :label="index == 0 ? $t('website.weight') : ''"
                        :prop="`servers.${index}.weight`"
                        :rules="rules.weight"
                    >
                        <el-input type="number" v-model.number="item.servers[index].weight"></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="4">
                    <el-form-item
                        :label="index == 0 ? $t('website.maxFails') : ''"
                        :prop="`servers.${index}.maxFails`"
                        :rules="rules.maxFails"
                    >
                        <el-input type="number" v-model.number="item.servers[index].maxFails"></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="4">
                    <el-form-item
                        :label="index == 0 ? $t('website.maxConns') : ''"
                        :prop="`servers.${index}.maxConns`"
                        :rules="rules.maxConns"
                    >
                        <el-input type="number" v-model.number="item.servers[index].maxConns"></el-input>
                    </el-form-item>
                </el-col>
                <el-col :span="3">
                    <el-form-item
                        :label="index == 0 ? $t('website.strategy') : ''"
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
                <el-col :span="3" v-if="index == 0">
                    <el-form-item :label="$t('commons.button.add') + $t('website.server')">
                        <el-button @click="addServer">
                            <el-icon><Plus /></el-icon>
                        </el-button>
                    </el-form-item>
                </el-col>
                <el-col :span="3" v-else>
                    <el-form-item>
                        <el-button @click="removeServer(index)" link type="primary">
                            <el-icon><Delete /></el-icon>
                        </el-button>
                    </el-form-item>
                </el-col>
            </el-row>
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
import { MsgSuccess } from '@/utils/message';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { getAlgorithms, getStatusStrategy } from '@/global/mimetype';
import { Website } from '@/api/interface/website';

const rules = ref<any>({
    name: [Rules.appName],
    algorithm: [Rules.requiredSelect],
    server: [Rules.requiredInput],
    weight: [checkNumberRange(0, 100)],
    servers: {
        type: Array,
    },
    maxFails: [checkNumberRange(1, 1000)],
    maxConns: [checkNumberRange(1, 1000)],
});

interface LoadBalanceOperate {
    websiteID: number;
    operate: string;
    upstream?: Website.NginxUpstream;
}

const lbForm = ref<FormInstance>();

const initServer = () => ({
    server: '',
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
            servers.push({
                server: server.server,
                weight: weight,
                maxFails: maxFails,
                maxConns: maxConns,
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

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) {
            return;
        }
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
        } finally {
            loading.value = false;
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
