<template>
    <div v-loading="loading">
        <div>
            <el-form-item :label="$t('website.enable')">
                <el-switch v-model="data.enable" @change="updateEnable"></el-switch>
            </el-form-item>
        </div>
        <LogFile :config="{ id: id, type: 'website', name: logName, colorMode: 'nginx' }" ref="logRef">
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
import { getWebsite, opWebsiteLog } from '@/api/modules/website';
import i18n from '@/lang';
import LogFile from '@/components/log/file/index.vue';
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
const logName = computed(() => {
    return props.logType;
});
const id = computed(() => {
    return props.id;
});
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
        logType: props.logType,
    };
    loading.value = true;
    opWebsiteLog(req)
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
    let log = props.logType === 'access.log' ? i18n.global.t('logs.websiteLog') : i18n.global.t('website.errLog');
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.clean'),
        names: [],
        msg: i18n.global.t('commons.msg.operatorHelper', [log, i18n.global.t('commons.button.clean')]),
        api: opWebsiteLog,
        params: { id: id.value, operate: 'delete', logType: props.logType },
    });
};

const get = async () => {
    try {
        const res = await getWebsite(props.id);
        if (props.logType === 'access.log') {
            data.value.enable = res.data.accessLog;
        }
        if (props.logType === 'error.log') {
            data.value.enable = res.data.errorLog;
        }
    } catch (error) {}
};

onMounted(() => {
    get();
});
</script>
