<template>
    <div v-loading="loading">
        <CodemirrorPro class="mt-5" v-model="content" :heightDiff="375"></CodemirrorPro>
        <div class="mt-5">
            <el-button type="primary" @click="submit()" :disabled="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { OperateSupervisorConfig } from '@/api/modules/host-tool';

let data = ref();
let content = ref('');
let loading = ref(false);

const submit = () => {
    loading.value = true;
    OperateSupervisorConfig({ type: 'supervisord', operate: 'set', content: content.value })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            getConfig();
        })
        .finally(() => {
            loading.value = false;
        });
};

const getConfig = async () => {
    const res = await OperateSupervisorConfig({ type: 'supervisord', operate: 'get' });
    data.value = res.data;
    content.value = data.value.content;
};

onMounted(() => {
    getConfig();
});
</script>
