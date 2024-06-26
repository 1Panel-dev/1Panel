<template>
    <div>
        <div class="app-status tool-status" v-if="data.isExist">
            <el-card>
                <div>
                    <el-tag effect="dark" type="success">Clamav</el-tag>
                    <el-tag round class="status-content" v-if="data.isActive" type="success">
                        {{ $t('commons.status.running') }}
                    </el-tag>
                    <el-tag round class="status-content" v-if="!data.isActive" type="info">
                        {{ $t('commons.status.stopped') }}
                    </el-tag>
                    <el-tag class="status-content">{{ $t('app.version') }}:{{ data.version }}</el-tag>
                    <span class="buttons">
                        <el-button type="primary" v-if="!data.isActive" link @click="onOperate('start')">
                            {{ $t('app.start') }}
                        </el-button>
                        <el-button type="primary" v-if="data.isActive" link @click="onOperate('stop')">
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
                    </span>
                </div>
            </el-card>
        </div>
        <LayoutContent :title="$t('toolbox.clam.clam')" :divider="true" v-if="!data.isExist" v-loading="loading">
            <template #main>
                <div class="app-warn">
                    <div>
                        <span v-if="!data.isExist">{{ $t('toolbox.clam.noClam') }}</span>
                        <span @click="toDoc()" v-if="!data.isExist">
                            <el-icon class="ml-2"><Position /></el-icon>
                            {{ $t('firewall.quickJump') }}
                        </span>
                        <div>
                            <img alt="" src="@/assets/images/no_app.svg" />
                        </div>
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

const data = ref({
    isExist: false,
    isActive: false,
    version: '',
});
const loading = ref(false);

const em = defineEmits(['setting', 'getStatus', 'update:loading', 'update:maskShow']);

const setting = () => {
    em('setting', true);
};

const toDoc = async () => {
    window.open('https://1panel.cn/docs/user_manual/toolbox/clam/', '_blank', 'noopener,noreferrer');
};

const onOperate = async (operation: string) => {
    em('update:maskShow', false);
    ElMessageBox.confirm(
        i18n.global.t('commons.msg.operatorHelper', [' Clamav ', i18n.global.t('app.' + operation)]),
        i18n.global.t('app.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    )
        .then(() => {
            em('update:loading', true);
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
