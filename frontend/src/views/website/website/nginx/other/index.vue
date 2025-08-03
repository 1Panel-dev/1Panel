<template>
    <div v-if="showPage" v-loading="loading">
        <el-text>{{ $t('nginx.defaultHttps') }}</el-text>
        <el-switch class="ml-2" v-model="enable" @change="changeStatus"></el-switch>
        <div>
            <el-text type="info" size="small">{{ $t('nginx.defaultHttpsHelper1') }}</el-text>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { getHttpsStatus, operateHttps } from '@/api/modules/nginx';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const showPage = ref(false);
const enable = ref(false);
const loading = ref(false);

const getStatus = async () => {
    try {
        const res = await getHttpsStatus();
        enable.value = res.data.https;
        showPage.value = true;
    } catch {}
};

const changeStatus = async () => {
    loading.value = true;
    try {
        await operateHttps({ operate: enable.value ? 'enable' : 'disable' });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch (error) {
        enable.value = !enable.value;
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    getStatus();
});
</script>
