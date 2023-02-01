<template>
    <div>
        <div class="a-card" v-if="data.isExist">
            <el-card>
                <div>
                    <el-tag effect="dark" type="success">{{ data.app }}</el-tag>
                    <Status class="status-content" :key="refresh" :status="data.status"></Status>
                    <el-tag class="status-content">{{ $t('app.version') }}:{{ data.version }}</el-tag>

                    <span class="buttons">
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
                    </span>
                </div>
            </el-card>
        </div>
        <div v-else>
            <div class="app-warn">
                <div>
                    <span>{{ $t('app.checkInstalledWarn', [data.app]) }}</span>
                </div>
                <el-link icon="Position" @click="goRouter('/apps')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </div>
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

<style lang="scss" scoped>
.a-card {
    font-size: 12px;

    .el-card {
        --el-card-padding: 12px;

        .buttons {
            margin-left: 100px;
        }
    }
}

body {
    margin: 0;
}

.status-content {
    margin-left: 50px;
}

.app-warn {
    text-align: center;
    margin-top: 200px;
    span {
        font-weight: 500;
        font-size: 16px;
        color: #bbbfc4;
    }
}
</style>
