<template>
    <div class="flx-center">
        <span v-if="props.footer">
            <el-button type="primary" link @click="toForum">
                <span>{{ $t('setting.forum') }}</span>
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link @click="toDoc">
                <span>{{ $t('setting.doc2') }}</span>
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link @click="toGithub">
                <span>{{ $t('setting.project') }}</span>
            </el-button>
            <el-divider direction="vertical" />
        </span>
        <el-button type="primary" link @click="toHalo">
            {{ isProductPro ? $t('license.pro') : $t('license.community') }}
        </el-button>
        <span class="version">{{ version }}</span>
        <el-badge is-dot style="margin-top: -3px" v-if="version !== 'Waiting' && globalStore.hasNewVersion">
            <el-button type="primary" link @click="onLoadUpgradeInfo">
                <span>({{ $t('setting.hasNewVersion') }})</span>
            </el-button>
        </el-badge>
        <el-button
            v-if="version !== 'Waiting' && !globalStore.hasNewVersion"
            type="primary"
            link
            @click="onLoadUpgradeInfo"
        >
            <span>({{ $t('setting.upgradeCheck') }})</span>
        </el-button>
        <el-tag v-if="version === 'Waiting'" round style="margin-left: 10px">{{ $t('setting.upgrading') }}</el-tag>
    </div>
    <el-drawer
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :key="refresh"
        v-model="drawerVisible"
        size="50%"
        append-to-body
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.upgrade')" :back="handleClose" />
        </template>
        <div class="panel-MdEditor">
            <el-alert :closable="false">
                <span class="line-height">{{ $t('setting.versionHelper') }}</span>
                <li class="line-height">{{ $t('setting.versionHelper1') }}</li>
                <li class="line-height">{{ $t('setting.versionHelper2') }}</li>
            </el-alert>
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
            <MdEditor v-model="upgradeInfo.releaseNote" previewOnly :theme="isDarkTheme ? 'dark' : 'light'" />
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);

const version = ref<string>('');
const isProductPro = ref();
const loading = ref(false);
const drawerVisible = ref(false);
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
    drawerVisible.value = false;
};

const toHalo = () => {
    window.open('https://www.lxware.cn/1panel' + '', '_blank', 'noopener,noreferrer');
};

const toDoc = () => {
    window.open('https://1panel.cn/docs/', '_blank', 'noopener,noreferrer');
};

const toForum = () => {
    window.open('https://bbs.fit2cloud.com/c/1p/7', '_blank');
};

const toGithub = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank', 'noopener,noreferrer');
};

const onLoadUpgradeInfo = async () => {
    loading.value = true;
    await loadUpgradeInfo()
        .then((res) => {
            loading.value = false;
            if (res.data.testVersion || res.data.newVersion || res.data.latestVersion) {
                upgradeInfo.value = res.data;
                drawerVisible.value = true;
                if (upgradeInfo.value.newVersion) {
                    upgradeVersion.value = upgradeInfo.value.newVersion;
                    return;
                }
                if (upgradeInfo.value.latestVersion) {
                    upgradeVersion.value = upgradeInfo.value.latestVersion;
                    return;
                }
                if (upgradeInfo.value.testVersion) {
                    upgradeVersion.value = upgradeInfo.value.testVersion;
                    return;
                }
            } else {
                MsgSuccess(i18n.global.t('setting.noUpgrade'));
                return;
            }
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
        globalStore.isOnRestart = true;
        drawerVisible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

onMounted(() => {
    isProductPro.value = globalStore.isProductPro;
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
        margin: 13px, 0;
        padding: 0;
        font-size: 16px;
    }
}
</style>
