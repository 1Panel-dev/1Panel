<template>
    <div>
        <div class="app-content" v-if="data.isExist">
            <el-card class="app-card">
                <el-row :gutter="20">
                    <el-col :xs="10" :sm="10" :md="10" :lg="10" :xl="6">
                        <el-tag effect="dark" type="success">{{ data.app }}</el-tag>
                        <Status class="status-content" :key="refresh" :status="data.status"></Status>
                        <el-tag class="status-content" type="primary">
                            {{ $t('app.version') }}:{{ data.version }}
                        </el-tag>
                    </el-col>
                    <el-col :xs="8" :sm="8" :md="8" :lg="6" :xl="4">
                        <el-button type="primary" v-if="data.status != 'Running'" link @click="onOperate('up')">
                            {{ $t('app.up') }}
                        </el-button>
                        <el-button type="primary" v-if="data.status === 'Running'" link @click="onOperate('down')">
                            {{ $t('app.down') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link @click="onOperate('restart')">
                            {{ $t('app.restart') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button
                            type="primary"
                            @click="setting"
                            link
                            :disabled="data.status !== 'Running' && data.app === 'OpenResty'"
                        >
                            {{ $t('commons.button.set') }}
                        </el-button>
                    </el-col>
                </el-row>
            </el-card>
        </div>
        <div v-else>
            <el-alert :closable="false" :title="$t('app.checkInstalledWarn', [data.app])" type="info">
                <el-link icon="Position" @click="goRouter('/apps')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </el-alert>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { CheckAppInstalled, InstalledOp } from '@/api/modules/app';
import router from '@/routers';
import { onMounted, reactive, ref } from 'vue';
import Status from '@/components/status/index.vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import i18n from '@/lang';

const props = defineProps({
    appKey: {
        type: String,
        default: 'nginx',
    },
});

let key = ref('');
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

const em = defineEmits(['setting', 'isExist', 'before', 'update:loading']);
const setting = () => {
    em('setting', false);
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

const onCheck = async () => {
    const res = await CheckAppInstalled(key.value);
    data.value = res.data;
    em('isExist', res.data);
    operateReq.installId = res.data.appInstallId;
    refresh.value++;
};

const onOperate = async (operation: string) => {
    operateReq.operate = operation;
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('app.' + operation)]),
        i18n.global.t('app.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        em('update:loading', true);
        em('before');
        InstalledOp(operateReq)
            .then(() => {
                em('update:loading', false);
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                onCheck();
            })
            .catch(() => {
                em('update:loading', false);
            });
    });
};

onMounted(() => {
    key.value = props.appKey;
    onCheck();
});
</script>

<style lang="scss">
.app-card {
    font-size: 14px;
    height: 60px;
}

.app-content {
    height: 50px;
}

body {
    margin: 0;
}

.status-content {
    margin-left: 50px;
}
</style>
