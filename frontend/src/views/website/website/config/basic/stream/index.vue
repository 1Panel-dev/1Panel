<template>
    <div>
        <el-form ref="streamFormRef" :model="form" label-width="120px" label-position="top">
            <el-row v-loading="loading">
                <el-col :span="12" :offset="1">
                    <el-form-item :label="$t('website.streamPorts')" :rules="Rules.requiredInput">
                        <el-input v-model="form.streamPorts" />
                    </el-form-item>
                    <el-form-item prop="udp">
                        <el-checkbox v-model="form.udp" :label="$t('website.udp')" size="large" />
                    </el-form-item>
                    <LoadBalanceForm ref="lbFormRef" v-model="form" :disabled="true" />
                    <el-form-item>
                        <el-button type="primary" @click="submit()" class="mt-2">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
    </div>
</template>

<script lang="ts" setup>
import LoadBalanceForm from '@/views/website/website/config/basic/load-balance/form/index.vue';

import { getWebsite, updateWebsiteStream } from '@/api/modules/website';
import { ref, onMounted } from 'vue';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const form = ref({
    streamPorts: '',
    servers: [],
    name: '',
    algorithm: '',
    websiteID: props.id,
    udp: false,
});
const streamFormRef = ref();
const lbFormRef = ref();
const loading = ref(false);

const submit = async () => {
    try {
        loading.value = true;
        const formValid = await streamFormRef.value?.validate();
        if (!formValid) return;
        const lbValid = await lbFormRef.value?.validate();
        if (!lbValid) return;
        await updateWebsiteStream(form.value);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } finally {
        loading.value = false;
    }
};

onMounted(async () => {
    const res = await getWebsite(props.id);
    form.value.streamPorts = res.data.streamPorts;
    form.value.name = res.data.primaryDomain;
    form.value.algorithm = res.data.algorithm == '' ? 'default' : res.data.algorithm;
    form.value.streamPorts = res.data.streamPorts;
    form.value.udp = res.data.udp;

    let servers = [];
    res.data?.servers?.forEach((server) => {
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
    form.value.servers = servers;
});
</script>
