<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="8" :lg="8" :xl="8">
            <el-form ref="nginxForm" label-position="right" label-width="180px" :model="form" :rules="rules">
                <el-form-item :label="$t('commons.table.status')" prop="enable">
                    <el-switch v-model="enable"></el-switch>
                    <span class="input-help">{{ $t('nginx.defaultHttpsHelper1') }}</span>
                </el-form-item>
                <el-form-item :label="$t('nginx.sslRejectHandshake')" prop="sslRejectHandshake">
                    <el-switch v-model="form.sslRejectHandshake"></el-switch>
                    <span class="input-help">{{ $t('nginx.sslRejectHandshakeHelper') }}</span>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="submit()" :disabled="loading">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { getHttpsStatus, operateHttps } from '@/api/modules/nginx';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const showPage = ref(false);
const enable = ref(false);
const loading = ref(false);
const form = reactive({
    operate: 'enable',
    sslRejectHandshake: false,
});

const getStatus = async () => {
    try {
        const res = await getHttpsStatus();
        enable.value = res.data.https;
        form.sslRejectHandshake = res.data.sslRejectHandshake;
        showPage.value = true;
    } catch {}
};

const submit = async () => {
    loading.value = true;
    try {
        form.operate = enable.value ? 'enable' : 'disable';
        await operateHttps(form);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch (error) {
        getStatus();
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    getStatus();
});
</script>
