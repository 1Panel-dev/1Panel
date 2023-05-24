<template>
    <div>
        <div class="app-status" v-if="data.isExist">
            <el-card>
                <div>
                    <el-tag effect="dark" type="success">{{ data.app }}</el-tag>
                    <Status class="status-content" :key="refresh" :status="data.status"></Status>
                    <el-tag class="status-content">{{ $t('app.version') }}:{{ data.version }}</el-tag>

                    <span class="buttons">
                        <el-button
                            type="primary"
                            v-if="data.status != 'Running'"
                            link
                            @click="onOperate('start')"
                            :disabled="data.status === 'Installing'"
                        >
                            {{ $t('app.start') }}
                        </el-button>
                        <el-button type="primary" v-if="data.status === 'Running'" link @click="onOperate('stop')">
                            {{ $t('app.stop') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button
                            type="primary"
                            link
                            :disabled="data.status === 'Installing'"
                            @click="onOperate('restart')"
                        >
                            {{ $t('app.restart') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button
                            type="primary"
                            @click="setting"
                            link
                            :disabled="
                                data.status === 'Installing' || (data.status !== 'Running' && data.app === 'OpenResty')
                            "
                        >
                            {{ $t('commons.button.set') }}
                        </el-button>
                    </span>
                </div>
            </el-card>
        </div>
        <div v-else>
            <LayoutContent :title="getTitle(key)" :divider="true">
                <template #main>
                    <div class="app-warn">
                        <div>
                            <span>{{ $t('app.checkInstalledWarn', [data.app]) }}</span>
                            <span @click="goRouter(key)">
                                <el-icon><Position /></el-icon>
                                {{ $t('database.goInstall') }}
                            </span>
                            <div>
                                <img src="@/assets/images/no_app.svg" />
                            </div>
                        </div>
                    </div>
                </template>
            </LayoutContent>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { CheckAppInstalled, InstalledOp } from '@/api/modules/app';
import router from '@/routers';
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

const em = defineEmits(['setting', 'isExist', 'before', 'update:loading', 'update:maskShow']);
const setting = () => {
    em('setting', false);
};

const goRouter = async (key: string) => {
    router.push({ name: 'AppDetail', params: { appKey: key } });
};

const onCheck = async () => {
    const res = await CheckAppInstalled(key.value);
    data.value = res.data;
    em('isExist', res.data);
    operateReq.installId = res.data.appInstallId;
    refresh.value++;
};

const onOperate = async (operation: string) => {
    em('update:maskShow', false);
    operateReq.operate = operation;
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('app.' + operation)]),
        i18n.global.t('app.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    )
        .then(() => {
            em('update:maskShow', true);
            em('update:loading', true);
            em('before');
            InstalledOp(operateReq)
                .then(() => {
                    em('update:loading', false);
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    onCheck();
                })
                .catch(() => {
                    em('update:loading', false);
                });
        })
        .catch(() => {
            em('update:maskShow', true);
        });
};

const getTitle = (key: string) => {
    switch (key) {
        case 'openresty':
            return i18n.global.t('website.website');
        case 'mysql':
            return 'MySQL ' + i18n.global.t('menu.database');
        case 'redis':
            return 'Redis ' + i18n.global.t('menu.database');
    }
};

onMounted(() => {
    key.value = props.appKey;
    onCheck();
});
</script>
