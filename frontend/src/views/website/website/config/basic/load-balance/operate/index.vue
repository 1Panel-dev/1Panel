<template>
    <DrawerPro
        v-model="open"
        @close="handleClose"
        size="large"
        :header="$t('commons.button.' + operate) + $t('website.loadBalance')"
        :resource="operate == 'create' ? '' : formData.name"
    >
        <LoadBalanceForm ref="lbFormRef" v-model="formData" :disabled="operate === 'edit' && disableName" />

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { createLoadBalance, updateLoadBalance } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Website } from '@/api/interface/website';
import LoadBalanceForm from '@/views/website/website/config/basic/load-balance/form/index.vue';

interface LoadBalanceOperate {
    websiteID: number;
    operate: string;
    upstream?: Website.NginxUpstream;
}

const lbFormRef = ref();
const open = ref(false);
const loading = ref(false);
const operate = ref('create');
const websiteID = ref(0);
const disableName = ref(false);

const formData = ref({
    name: '',
    algorithm: 'default',
    servers: [
        {
            server: '',
            weight: undefined,
            maxFails: undefined,
            maxConns: undefined,
            failTimeout: undefined,
            failTimeoutUnit: 's',
            flag: '',
        },
    ],
});

const em = defineEmits(['close']);

const handleClose = () => {
    lbFormRef.value?.resetFields();
    open.value = false;
    em('close', false);
};

const acceptParams = async (req: LoadBalanceOperate) => {
    websiteID.value = req.websiteID;
    operate.value = req.operate;

    if (req.operate == 'edit') {
        disableName.value = true;
        formData.value.name = req.upstream?.name || '';
        formData.value.algorithm = req.upstream?.algorithm || 'default';

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
        formData.value.servers = servers;
    } else {
        disableName.value = false;
        formData.value.name = '';
        formData.value.algorithm = 'default';
        formData.value.servers = [
            {
                server: '',
                weight: undefined,
                maxFails: undefined,
                maxConns: undefined,
                failTimeout: undefined,
                failTimeoutUnit: 's',
                flag: '',
            },
        ];
    }
    open.value = true;
};

const handleServers = () => {
    for (const server of formData.value.servers) {
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
    for (const server of formData.value.servers) {
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

const submit = async () => {
    try {
        const valid = await lbFormRef.value?.validate();
        if (!valid) return;

        if (formData.value.algorithm == 'ip_hash') {
            for (const server of formData.value.servers) {
                if (server.flag == 'backup') {
                    MsgError(i18n.global.t('website.ipHashBackupErr'));
                    return;
                }
            }
        }

        handleServers();
        loading.value = true;

        const submitData = {
            websiteID: websiteID.value,
            ...formData.value,
        };

        if (operate.value === 'edit') {
            await updateLoadBalance(submitData);
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        } else {
            await createLoadBalance(submitData);
            MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
        }
        handleClose();
    } catch (error) {
        rollBackServers();
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
