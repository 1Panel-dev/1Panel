<template>
    <div>
        <div class="app-status tool-status" v-if="data.isExist">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-tag effect="dark" type="success">{{ 'Supervisor' }}</el-tag>
                        <Status :key="data.status" :status="data.status"></Status>
                        <el-tag>{{ $t('app.version') }}{{ $t('commons.colon') }}{{ data.version }}</el-tag>
                    </div>
                    <div class="mt-0.5" v-if="!data.init">
                        <el-button type="primary" v-if="data.status != 'running'" link @click="onOperate('start')">
                            {{ $t('app.start') }}
                        </el-button>
                        <el-button type="primary" v-if="data.status == 'running'" link @click="onOperate('stop')">
                            {{ $t('app.stop') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link @click="onOperate('restart')">
                            {{ $t('app.restart') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link @click="setting">
                            {{ $t('commons.button.set') }}
                        </el-button>
                    </div>
                    <span class="buttons" v-else>
                        <el-button type="primary" link @click="init">
                            {{ $t('commons.button.init') }}
                        </el-button>
                    </span>
                </div>
            </el-card>
        </div>
        <LayoutContent
            :title="$t('tool.supervisor.list')"
            :divider="true"
            v-if="!data.isExist || !data.ctlExist || data.init"
            v-loading="loading"
        >
            <template #main>
                <div class="app-warn">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <span v-if="!data.isExist">{{ $t('tool.supervisor.notSupport') }}</span>
                        <span v-else-if="!data.ctlExist">{{ $t('tool.supervisor.notSupportCrl') }}</span>
                        <span v-else-if="data.init">{{ $t('tool.supervisor.initHelper') }}</span>
                        <span
                            @click="toDoc()"
                            v-if="!data.isExist || !data.ctlExist"
                            class="flex items-center justify-center gap-0.5"
                        >
                            <el-icon><Position /></el-icon>
                            {{ $t('firewall.quickJump') }}
                        </span>
                    </div>
                    <div>
                        <img alt="" src="@/assets/images/no_app.svg" />
                    </div>
                </div>
            </template>
        </LayoutContent>
        <InitPage ref="initRef" @close="getStatus"></InitPage>
    </div>
</template>
<script lang="ts" setup>
import { GetSupervisorStatus, OperateSupervisor } from '@/api/modules/host-tool';
import { onMounted, reactive, ref } from 'vue';
import Status from '@/components/status/index.vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { HostTool } from '@/api/interface/host-tool';
import InitPage from './init/index.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
let operateReq = reactive({
    installId: 0,
    operate: '',
});
const initRef = ref();
const data = ref({
    isExist: false,
    version: '',
    status: 'running',
    init: false,
    configPath: '',
    ctlExist: false,
    serviceName: '',
});
const loading = ref(false);

const em = defineEmits(['setting', 'getStatus', 'update:loading', 'update:maskShow']);

const setting = () => {
    em('setting', true);
};

const toDoc = async () => {
    window.open(globalStore.docsUrl + '/user_manual/toolbox/supervisor/', '_blank', 'noopener,noreferrer');
};

const init = async () => {
    initRef.value.acceptParams(data.value.configPath, data.value.serviceName);
};

const onOperate = async (operation: string) => {
    em('update:maskShow', false);
    operateReq.operate = operation;
    ElMessageBox.confirm(
        i18n.global.t('tool.supervisor.operatorHelper', ['Supervisor', i18n.global.t('app.' + operation)]),
        i18n.global.t('app.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    )
        .then(() => {
            em('update:loading', true);
            OperateSupervisor(operation)
                .then(() => {
                    em('update:maskShow', true);
                    getStatus();
                    em('update:loading', false);
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .catch(() => {
                    em('update:loading', false);
                });
        })
        .catch(() => {
            em('update:maskShow', true);
        });
};

const getStatus = async () => {
    try {
        loading.value = true;
        em('update:loading', true);
        const res = await GetSupervisorStatus();
        if (res.data.config) {
            data.value = res.data.config as HostTool.Supersivor;
        }

        const status = {
            isExist: data.value.isExist && data.value.ctlExist,
            isRunning: data.value.status === 'running',
            init: data.value.init,
        };
        em('getStatus', status);
    } catch (error) {}
    em('update:loading', false);
    loading.value = false;
};

onMounted(() => {
    getStatus();
});
</script>

<style lang="scss" scoped>
.tool-status {
    margin-top: 20px;
}
</style>
