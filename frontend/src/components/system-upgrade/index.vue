<template>
    <div>
        <span v-if="props.footer">
            <el-button type="primary" link @click="toForum">
                <span>{{ $t('setting.forum') }}</span>
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link @click="toDoc">
                <span>{{ $t('setting.doc2') }}</span>
            </el-button>
            <el-divider direction="vertical" />
        </span>
        <span class="version">{{ $t('setting.currentVersion') + version }}</span>
        <el-badge
            is-dot
            class="item"
            v-if="version !== 'Waiting' && globalStore.hasNewVersion"
            style="margin-top: -6px"
        >
            <el-button type="primary" link @click="onLoadUpgradeInfo">
                <span style="font-size: 14px">（{{ $t('setting.hasNewVersion') }}）</span>
            </el-button>
        </el-badge>
        <el-button
            v-if="version !== 'Waiting' && !globalStore.hasNewVersion"
            style="margin-top: -2px"
            type="primary"
            link
            @click="onLoadUpgradeInfo"
        >
            （{{ $t('setting.upgradeCheck') }}）
        </el-button>
        <el-tag v-if="version === 'Waiting'" round style="margin-left: 10px">{{ $t('setting.upgrading') }}</el-tag>
    </div>
    <el-drawer :close-on-click-modal="false" :key="refresh" v-model="drawerVisiable" size="50%" append-to-body>
        <template #header>
            <DrawerHeader :header="$t('commons.button.upgrade')" :back="handleClose" />
        </template>
        <div class="panel-MdEditor">
            <el-alert :closable="false">
                {{ $t('setting.versionHelper') }}
                <li>{{ $t('setting.versionHelper1') }}</li>
                <li>{{ $t('setting.versionHelper2') }}</li>
            </el-alert>
            <div class="default-theme">
                <h2 class="inline-block">{{ $t('app.version') }}</h2>
            </div>
            <el-radio-group class="inline-block tag" v-model="upgradeVersion" @change="changeOption">
                <el-radio v-if="upgradeInfo.newVersion" :label="upgradeInfo.newVersion">
                    {{ upgradeInfo.newVersion }} {{ $t('setting.newVersion') }}
                </el-radio>
                <el-radio :label="upgradeInfo.latestVersion">
                    {{ upgradeInfo.latestVersion }} {{ $t('setting.latestVersion') }}
                </el-radio>
            </el-radio-group>
            <MdEditor
                v-model="upgradeInfo.releaseNote"
                previewOnly
                :theme="globalStore.$state.themeConfig.theme === 'dark' ? 'dark' : 'light'"
            />
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
import DrawerHeader from '@/components/drawer-header/index.vue';
import { getSettingInfo, loadReleaseNotes, loadUpgradeInfo, upgrade } from '@/api/modules/setting';
import MdEditor from 'md-editor-v3';
import i18n from '@/lang';
import 'md-editor-v3/lib/style.css';
import { MsgSuccess } from '@/utils/message';
import { onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';
import { ElMessageBox } from 'element-plus';
const globalStore = GlobalStore();

const version = ref();
const loading = ref(false);
const drawerVisiable = ref(false);
const upgradeInfo = ref();
const refresh = ref();
const upgradeVersion = ref();
const props = defineProps({
    footer: {
        type: Boolean,
        default: false,
    },
});

const search = async () => {
    const res = await getSettingInfo();
    version.value = res.data.systemVersion;
};

const handleClose = () => {
    drawerVisiable.value = false;
};

const toDoc = () => {
    window.open('https://1panel.cn/docs/', '_blank');
};

const toForum = () => {
    window.open('https://bbs.fit2cloud.com/c/1p/7', '_blank');
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
            upgradeVersion.value = upgradeInfo.value.newVersion || upgradeInfo.value.latestVersion;
            drawerVisiable.value = true;
        })
        .catch(() => {
            loading.value = false;
        });
};

const changeOption = async () => {
    const res = await loadReleaseNotes(upgradeVersion.value);
    upgradeInfo.value.releaseNote = res.data;
};

const onUpgrade = async () => {
    ElMessageBox.confirm(i18n.global.t('setting.upgradeHelper', i18n.global.t('commons.button.upgrade')), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        globalStore.isLoading = true;
        await upgrade(upgradeVersion.value);
        drawerVisiable.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
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
