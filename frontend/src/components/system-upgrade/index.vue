<template>
    <div>
        <span class="version">{{ version }}</span>
        <el-button v-if="version !== 'Waiting'" type="primary" link @click="onLoadUpgradeInfo">
            {{ $t('setting.upgradeCheck') }}
        </el-button>
        <el-tag v-else round style="margin-left: 10px">{{ $t('setting.upgrading') }}</el-tag>
    </div>

    <el-drawer :close-on-click-modal="false" :key="refresh" v-model="drawerVisiable" size="50%" append-to-body>
        <template #header>
            <DrawerHeader :header="$t('setting.upgrade')" :back="handleClose" />
        </template>
        <div class="panel-MdEditor">
            <div class="default-theme">
                <h2 class="inline-block">{{ $t('setting.newVersion') }}</h2>
                <el-tag class="inline-block tag">{{ upgradeInfo.newVersion }}</el-tag>
            </div>
            <MdEditor v-model="upgradeInfo.releaseNote" previewOnly />
        </div>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onUpgrade">{{ $t('setting.upgradeNow') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>
<script setup lang="ts">
import { getSettingInfo, loadUpgradeInfo, upgrade } from '@/api/modules/setting';
import MdEditor from 'md-editor-v3';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { onMounted, ref } from 'vue';

const version = ref();
let loading = ref(false);
const drawerVisiable = ref(false);
const upgradeInfo = ref();
const refresh = ref();

const search = async () => {
    const res = await getSettingInfo();
    version.value = res.data.systemVersion;
};

const handleClose = () => {
    drawerVisiable.value = false;
};

const onLoadUpgradeInfo = async () => {
    loading.value = true;
    await loadUpgradeInfo()
        .then((res) => {
            loading.value = false;

            if (!res.data) {
                MsgSuccess(i18n.global.t('setting.noUpgrade'));
                return;
            }
            upgradeInfo.value = res.data;
            drawerVisiable.value = true;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onUpgrade = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.upgradeHelper', i18n.global.t('setting.upgrade')), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(() => {
        loading.value = true;
        upgrade(upgradeInfo.value.newVersion)
            .then(() => {
                loading.value = false;
                drawerVisiable.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
.version {
    font-size: 14px;
    color: #858585;
    text-decoration: none;
    letter-spacing: 0.5px;
}
.panel-MdEditor {
    height: calc(100vh - 330px);
    margin-left: 70px;
    .tag {
        margin-left: 20px;
        margin-top: -6px;
        vertical-align: middle;
    }
    :deep(.md-editor-preview) {
        font-size: 14px;
    }
    :deep(.default-theme h2) {
        margin: 13px 0;
        padding: 0;
        font-size: 16px;
    }
}
</style>
