<template>
    <div>
        <div class="app-status" v-if="data.isExist">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-tag effect="dark" type="success">ClamAV</el-tag>
                        <el-tag round v-if="data.isActive" type="success">
                            {{ $t('commons.status.running') }}
                        </el-tag>
                        <el-tag round v-if="!data.isActive" type="info">
                            {{ $t('commons.status.stopped') }}
                        </el-tag>
                        <el-tag class="w-24">{{ $t('app.version') }}{{ $t('commons.colon') }}{{ data.version }}</el-tag>
                    </div>
                    <div class="mt-0.5">
                        <el-button type="primary" v-if="!data.isActive" link @click="onOperate('ClamAV', 'start')">
                            {{ $t('app.start') }}
                        </el-button>
                        <el-button type="primary" v-if="data.isActive" link @click="onOperate('ClamAV', 'stop')">
                            {{ $t('app.stop') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link @click="onOperate('ClamAV', 'restart')">
                            {{ $t('app.restart') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link @click="setting">
                            {{ $t('commons.button.set') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" v-if="showFresh" link @click="changeShow(false)">
                            {{ $t('toolbox.clam.hideFresh') }}
                        </el-button>
                        <el-button type="primary" v-if="!showFresh" link @click="changeShow(true)">
                            {{ $t('toolbox.clam.showFresh') }}
                        </el-button>
                    </div>
                </div>
                <div v-if="showFresh" class="flex w-full flex-col gap-4 md:flex-row mt-5">
                    <div class="flex flex-wrap gap-4">
                        <el-tag class="w-16" effect="dark" type="success">FreshClam</el-tag>
                        <el-tag round v-if="data.freshIsActive" type="success">
                            {{ $t('commons.status.running') }}
                        </el-tag>
                        <el-tag round v-if="!data.freshIsActive" type="info">
                            {{ $t('commons.status.stopped') }}
                        </el-tag>
                        <el-tag class="w-24">{{ $t('app.version') }}:{{ data.freshVersion }}</el-tag>
                    </div>
                    <div class="mt-0.5">
                        <el-button
                            type="primary"
                            v-if="!data.freshIsActive"
                            link
                            @click="onOperate('FreshClam', 'start')"
                        >
                            {{ $t('app.start') }}
                        </el-button>
                        <el-button
                            type="primary"
                            v-if="data.freshIsActive"
                            link
                            @click="onOperate('FreshClam', 'stop')"
                        >
                            {{ $t('app.stop') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" link @click="onOperate('FreshClam', 'restart')">
                            {{ $t('app.restart') }}
                        </el-button>
                    </div>
                </div>
            </el-card>
        </div>
        <LayoutContent :title="$t('toolbox.clam.clam')" :divider="true" v-if="!data.isExist" v-loading="loading">
            <template #main>
                <div class="app-warn">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <span v-if="!data.isExist">{{ $t('toolbox.clam.noClam') }}</span>
                        <span @click="toDoc()" v-if="!data.isExist" class="flex items-center justify-center gap-0.5">
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
    </div>
</template>
<script lang="ts" setup>
import { searchClamBaseInfo, updateClamBaseInfo } from '@/api/modules/toolbox';
import { onMounted, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

const data = ref({
    isExist: false,
    isActive: false,
    version: '',

    freshIsExist: false,
    freshIsActive: false,
    freshVersion: '',
});
const loading = ref(false);
const showFresh = ref(localStorage.getItem('clam-fresh-show') !== 'hide');

const em = defineEmits(['setting', 'getStatus', 'update:loading', 'update:maskShow']);

const setting = () => {
    em('setting', true);
};

const toDoc = async () => {
    window.open(globalStore.docsUrl + '/user_manual/toolbox/clam/', '_blank', 'noopener,noreferrer');
};

const changeShow = (val: boolean) => {
    showFresh.value = val;
    localStorage.setItem('clam-fresh-show', showFresh.value ? 'show' : 'hide');
};

const onOperate = async (service: string, operation: string) => {
    em('update:maskShow', false);
    ElMessageBox.confirm(
        i18n.global.t('commons.msg.operatorHelper', [' ' + service + ' ', i18n.global.t('app.' + operation)]),
        i18n.global.t('app.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    )
        .then(() => {
            em('update:loading', true);
            if (service === 'FreshClam') {
                operation = 'fresh-' + operation;
            }
            updateClamBaseInfo(operation)
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
        const res = await searchClamBaseInfo();
        data.value = res.data;
        const status = {
            isExist: data.value.isExist,
            isRunning: data.value.isActive,
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
