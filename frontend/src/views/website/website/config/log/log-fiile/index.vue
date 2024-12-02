<template>
    <div v-loading="loading">
        <div>
            <el-form-item :label="$t('website.enable')">
                <el-switch v-model="data.enable" @change="updateEnable"></el-switch>
            </el-form-item>
        </div>
        <LogFile :config="{ id: id, type: 'website', name: logType }" :style="style" ref="logRef">
            <template #button>
                <el-button @click="cleanLog" icon="Delete">
                    {{ $t('commons.button.clean') }}
                </el-button>
            </template>
        </LogFile>
    </div>
    <OpDialog ref="opRef" @search="clearLog" />
</template>
<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { GetWebsite, OpWebsiteLog } from '@/api/modules/website';
import i18n from '@/lang';
import LogFile from '@/components/log-file/index.vue';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    logType: {
        type: String,
        default: '',
    },
    id: {
        type: Number,
        default: 0,
    },
});
const logType = computed(() => {
    return props.logType;
});
const id = computed(() => {
    return props.id;
});
const style = ref('height: calc(100vh - 402px); width: 100%; min-height: 300px');
const loading = ref(false);
const data = ref({
    enable: false,
    content: '',
    path: '',
});
const opRef = ref();
const logRef = ref();

const updateEnable = () => {
    const operate = data.value.enable ? 'enable' : 'disable';
    const req = {
        id: id.value,
        operate: operate,
        logType: logType.value,
    };
    loading.value = true;
    OpWebsiteLog(req)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const clearLog = () => {
    logRef.value.clearLog();
};

const cleanLog = async () => {
    let log = logType.value === 'access.log' ? i18n.global.t('website.accessLog') : i18n.global.t('website.errLog');
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.clean'),
        names: [],
        msg: i18n.global.t('commons.msg.operatorHelper', [log, i18n.global.t('commons.button.clean')]),
        api: OpWebsiteLog,
        params: { id: id.value, operate: 'delete', logType: logType.value },
    });
};

const getWebsite = async () => {
    try {
        const res = await GetWebsite(props.id);
        if (props.logType === 'access.log') {
            data.value.enable = res.data.accessLog;
        }
        if (props.logType === 'error.log') {
            data.value.enable = res.data.errorLog;
        }
    } catch (error) {}
};

onMounted(() => {
    getWebsite();
});
</script>
