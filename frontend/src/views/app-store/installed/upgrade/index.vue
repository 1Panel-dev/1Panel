<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="30%">
        <template #header>
            <Header
                :header="$t('commons.button.' + operateReq.operate)"
                :resource="resourceName"
                :back="handleClose"
            ></Header>
        </template>
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form
                    @submit.prevent
                    ref="updateRef"
                    :rules="rules"
                    label-position="top"
                    :model="operateReq"
                    v-loading="loading"
                >
                    <el-form-item :label="$t('app.versionSelect')" prop="detailId">
                        <el-select v-model="operateReq.detailId">
                            <el-option
                                v-for="(version, index) in versions"
                                :key="index"
                                :value="version.detailId"
                                :label="version.version"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item prop="backup" v-if="operateReq.operate === 'upgrade'">
                        <el-checkbox v-model="operateReq.backup" :label="$t('app.backupApp')" />
                        <span class="input-help">{{ $t('app.backupAppHelper') }}</span>
                    </el-form-item>
                    <el-form-item pro="pullImage" v-if="operateReq.operate === 'upgrade'">
                        <el-checkbox v-model="operateReq.pullImage" :label="$t('container.forcePull')" size="large" />
                        <span class="input-help">{{ $t('container.forcePullHelper') }}</span>
                    </el-form-item>
                </el-form>
            </el-col>
            <el-col :span="22" :offset="1">
                <div class="descriptions">
                    <el-descriptions direction="vertical">
                        <el-descriptions-item>
                            <el-link @click="toLink(app.website)">
                                <el-icon><OfficeBuilding /></el-icon>
                                <span>{{ $t('app.appOfficeWebsite') }}</span>
                            </el-link>
                        </el-descriptions-item>
                        <el-descriptions-item>
                            <el-link @click="toLink(app.document)">
                                <el-icon><Document /></el-icon>
                                <span>{{ $t('app.document') }}</span>
                            </el-link>
                        </el-descriptions-item>
                        <el-descriptions-item>
                            <el-link @click="toLink(app.github)">
                                <el-icon><Link /></el-icon>
                                <span>{{ $t('app.github') }}</span>
                            </el-link>
                        </el-descriptions-item>
                    </el-descriptions>
                </div>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onOperate" :disabled="versions == null || loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { GetAppUpdateVersions, IgnoreUpgrade, InstalledOp } from '@/api/modules/app';
import i18n from '@/lang';
import { ElMessageBox, FormInstance } from 'element-plus';
import { reactive, ref, onBeforeUnmount } from 'vue';
import Header from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import bus from '../../bus';

const updateRef = ref<FormInstance>();
const open = ref(false);
const loading = ref(false);
const versions = ref<App.VersionDetail[]>();
const operateReq = reactive({
    detailId: 0,
    operate: 'upgrade',
    installId: 0,
    backup: true,
    pullImage: true,
});
const resourceName = ref('');
const rules = ref<any>({
    detailId: [Rules.requiredSelect],
});
const app = ref();
const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', open);
};

const toLink = (link: string) => {
    window.open(link, '_blank');
};

const acceptParams = (id: number, name: string, op: string, appDetail: App.AppDetail) => {
    operateReq.installId = id;
    operateReq.operate = op;
    resourceName.value = name;
    app.value = appDetail;
    GetAppUpdateVersions(id).then((res) => {
        versions.value = res.data;
        if (res.data != null && res.data.length > 0) {
            operateReq.detailId = res.data[0].detailId;
        }
        open.value = true;
    });
};

const operate = async () => {
    loading.value = true;
    if (operateReq.operate === 'upgrade') {
        await InstalledOp(operateReq)
            .then(() => {
                MsgSuccess(i18n.global.t('app.upgradeStart'));
                bus.emit('upgrade', true);
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    } else {
        await IgnoreUpgrade(operateReq)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                bus.emit('upgrade', true);
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    }
};

const onOperate = async () => {
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.button.' + operateReq.operate)]),
        i18n.global.t('commons.button.upgrade'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        operate();
    });
};

onBeforeUnmount(() => {
    bus.off('upgrade');
});

defineExpose({
    acceptParams,
});
</script>
