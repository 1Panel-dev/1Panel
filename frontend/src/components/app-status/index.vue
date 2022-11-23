<template>
    <div>
        <div class="app-content" v-if="data.isExist">
            <el-card class="app-card" v-loading="loading">
                <el-row :gutter="20">
                    <el-col :span="1">
                        <div>
                            <el-tag effect="dark" type="success">{{ data.app }}</el-tag>
                        </div>
                    </el-col>
                    <el-col :span="2">
                        <div>
                            {{ $t('app.version') }}:
                            <el-tag type="info">{{ data.version }}</el-tag>
                        </div>
                    </el-col>
                    <el-col :span="2">
                        <div>
                            {{ $t('commons.table.status') }}:
                            <el-tag type="success">{{ data.status }}</el-tag>
                        </div>
                    </el-col>
                    <el-col :span="4">
                        <div>
                            {{ $t('website.lastBackupAt') }}:
                            <el-tag v-if="data.lastBackupAt != ''" type="info">{{ data.lastBackupAt }}</el-tag>
                            <span else>{{ $t('website.null') }}</span>
                        </div>
                    </el-col>
                    <el-col :span="6">
                        <el-button type="primary" link @click="onOperate('restart')">{{ $t('app.restart') }}</el-button>
                        <el-button type="primary" link @click="setting">{{ $t('commons.button.set') }}</el-button>
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
import i18n from '@/lang';
import router from '@/routers';
import { ElMessage, ElMessageBox } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';

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
let loading = ref(false);
let operateReq = reactive({
    installId: 0,
    operate: '',
});

const em = defineEmits(['setting', 'isExist']);
const setting = () => {
    em('setting', false);
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

const onCheck = async () => {
    loading.value = true;
    const res = await CheckAppInstalled(key.value);
    data.value = res.data;
    em('isExist', res.data);
    operateReq.installId = res.data.appInstallId;
    loading.value = false;
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
        loading.value = true;
        InstalledOp(operateReq)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                onCheck();
            })
            .finally(() => {
                loading.value = false;
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
</style>
