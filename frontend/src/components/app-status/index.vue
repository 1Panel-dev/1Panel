<template>
    <div class="app-status" v-if="data.isExist">
        <el-card>
            <div class="flex w-full flex-col gap-4 md:flex-row">
                <div class="flex flex-wrap gap-4 ml-3">
                    <el-tag effect="dark" type="success">{{ data.app }}</el-tag>
                    <Status class="mt-0.5" :key="refresh" :status="data.status"></Status>
                    <el-tag>{{ $t('app.version') }}{{ data.version }}</el-tag>
                </div>

                <div class="mt-0.5">
                    <el-button
                        type="primary"
                        v-if="data.status != 'Running'"
                        link
                        @click="onOperate('start')"
                        :disabled="data.status === 'Installing'"
                    >
                        {{ $t('commons.operate.start') }}
                    </el-button>
                    <el-button type="primary" v-if="data.status === 'Running'" link @click="onOperate('stop')">
                        {{ $t('commons.operate.stop') }}
                    </el-button>
                    <el-divider direction="vertical" />
                    <el-button
                        type="primary"
                        link
                        :disabled="data.status === 'Installing'"
                        @click="onOperate('restart')"
                    >
                        {{ $t('commons.operate.restart') }}
                    </el-button>
                    <el-divider v-if="!hideSetting" direction="vertical" />
                    <el-button
                        type="primary"
                        link
                        v-if="data.app === 'OpenResty'"
                        @click="onOperate('reload')"
                        :disabled="data.status !== 'Running'"
                    >
                        {{ $t('commons.operate.reload') }}
                    </el-button>
                    <el-divider v-if="data.app === 'OpenResty'" direction="vertical" />
                    <el-button
                        v-if="!hideSetting"
                        type="primary"
                        @click="setting"
                        link
                        :disabled="data.status === 'Installing'"
                    >
                        {{ $t('commons.button.set') }}
                    </el-button>
                </div>
                <div class="ml-5" v-if="key === 'openresty' && (httpPort != 80 || httpsPort != 443)">
                    <el-tooltip
                        effect="dark"
                        :content="$t('website.openrestyHelper', [httpPort, httpsPort])"
                        placement="top-start"
                    >
                        <el-alert
                            :title="$t('app.checkTitle')"
                            :closable="false"
                            center
                            type="warning"
                            show-icon
                            class="h-6 check-title"
                        />
                    </el-tooltip>
                </div>
            </div>
        </el-card>
    </div>
</template>
<script lang="ts" setup>
import { checkAppInstalled, installedOp } from '@/api/modules/app';
import { onMounted, reactive, ref } from 'vue';
import Status from '@/components/status/index.vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    appKey: {
        type: String,
        default: 'openresty',
    },
    appName: {
        type: String,
        default: '',
    },
    hideSetting: {
        type: Boolean,
        default: false,
    },
});

let key = ref('');
let name = ref('');

let data = ref({
    app: '',
    version: '',
    status: '',
    lastBackupAt: '',
    appInstallId: 0,
    isExist: false,
    containerName: '',
});
let operateReq = reactive({
    installId: 0,
    operate: '',
});
let refresh = ref(1);
const httpPort = ref(0);
const httpsPort = ref(0);

const em = defineEmits([
    'setting',
    'isExist',
    'before',
    'after',
    'update:loading',
    'update:maskShow',
    'update:appInstallID',
]);
const setting = () => {
    em('setting', false);
};

const onCheck = async (key: any, name: any) => {
    await checkAppInstalled(key, name)
        .then((res) => {
            data.value = res.data;
            em('isExist', res.data);
            em('update:maskShow', res.data.status !== 'Running');
            operateReq.installId = res.data.appInstallId;
            em('update:appInstallID', res.data.appInstallId);
            httpPort.value = res.data.httpPort;
            httpsPort.value = res.data.httpsPort;
            refresh.value++;
        })
        .catch(() => {
            em('isExist', false);
            refresh.value++;
        });
};

const onOperate = async (operation: string) => {
    operateReq.operate = operation;
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.operate.' + operation)]),
        i18n.global.t('commons.operate.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        em('update:maskShow', true);
        em('update:loading', true);
        em('before');
        installedOp(operateReq)
            .then(() => {
                em('update:loading', false);
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                onCheck(key.value, name.value);
                em('after');
            })
            .catch(() => {
                em('update:loading', false);
            });
    });
};

onMounted(() => {
    key.value = props.appKey;
    name.value = props.appName;
    onCheck(key.value, name.value);
});

defineExpose({
    onCheck,
});
</script>
<style scoped lang="scss">
.check-title {
    color: var(--el-color-warning);
    border: 1px solid var(--el-color-warning);
    background-color: transparent;
    padding: 8px 8px;
    width: 70px;
}
</style>
