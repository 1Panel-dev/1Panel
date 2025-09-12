<template>
    <DrawerPro v-model="drawerVisible" :header="$t('commons.button.upgrade')" @close="handleClose" size="large">
        <div class="panel-MdEditor">
            <div class="default-theme" style="margin-left: 20px">
                <h2 class="inline-block">{{ $t('app.version') }}</h2>
            </div>
            <el-radio-group class="inline-block tag" v-model="upgradeVersion" @change="changeOption">
                <el-radio v-if="upgradeInfo.newVersion" :value="upgradeInfo.newVersion">
                    {{ upgradeInfo.newVersion }}
                </el-radio>
                <el-radio v-if="upgradeInfo.latestVersion" :value="upgradeInfo.latestVersion">
                    {{ upgradeInfo.latestVersion }}
                </el-radio>
                <el-radio v-if="upgradeInfo.testVersion" :value="upgradeInfo.testVersion">
                    {{ upgradeInfo.testVersion }}
                </el-radio>
            </el-radio-group>
            <MdEditor
                v-loading="loading"
                v-model="upgradeInfo.releaseNote"
                previewOnly
                :theme="isDarkTheme ? 'dark' : 'light'"
            />
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onUpgrade">{{ $t('setting.upgradeNow') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { loadReleaseNotes, upgrade } from '@/api/modules/setting';
import MdEditor from 'md-editor-v3';
import i18n from '@/lang';
import 'md-editor-v3/lib/style.css';
import { MsgSuccess } from '@/utils/message';
import { ref } from 'vue';
import { GlobalStore } from '@/store';
import { ElMessageBox } from 'element-plus';
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);

const drawerVisible = ref(false);
const upgradeInfo = ref();
const loading = ref();
const upgradeVersion = ref();

interface DialogProps {
    upgradeInfo: number;
    upgradeVersion: string;
}
const acceptParams = (params: DialogProps): void => {
    upgradeInfo.value = params.upgradeInfo;
    upgradeVersion.value = params.upgradeVersion;
    drawerVisible.value = true;
};

const emit = defineEmits(['search']);

const handleClose = () => {
    drawerVisible.value = false;
};

const changeOption = async () => {
    loading.value = true;
    await loadReleaseNotes(upgradeVersion.value)
        .then((res) => {
            loading.value = false;
            upgradeInfo.value.releaseNote = res.data;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onUpgrade = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.upgradeHelper', i18n.global.t('commons.button.upgrade')), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        await upgrade(upgradeVersion.value);
        globalStore.isLoading = true;
        globalStore.isOnRestart = true;
        drawerVisible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
    });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.line-height {
    line-height: 25px;
}
.panel-MdEditor {
    height: calc(100vh - 330px);
    .tag {
        margin-top: -6px;
        margin-left: 20px;
        vertical-align: middle;
    }
    :deep(.md-editor-preview) {
        font-size: 14px;
    }
    :deep(.default-theme h2) {
        color: var(--el-color-primary);
        margin: 13px 0;
        padding: 0;
        font-size: 16px;
    }
}
:deep(.md-editor-dark) {
    background-color: var(--panel-main-bg-color-9);
}
</style>
