<template>
    <div class="flx-center">
        <span v-if="props.footer">
            <el-button type="primary" link @click="toForum">
                <span class="font-normal">{{ $t('setting.forum') }}</span>
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link @click="toDoc">
                <span class="font-normal">{{ $t('setting.doc2') }}</span>
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link @click="toGithub">
                <span class="font-normal">{{ $t('setting.project') }}</span>
            </el-button>
            <el-divider direction="vertical" />
        </span>
        <el-button type="primary" link @click="toHalo">
            <span class="font-normal">{{ isMasterProductPro ? $t('license.pro') : $t('license.community') }}</span>
        </el-button>
        <span class="version">{{ version }}</span>
        <el-badge is-dot style="margin-top: -3px" v-if="version !== 'Waiting' && globalStore.hasNewVersion">
            <el-button type="primary" link @click="onLoadUpgradeInfo">
                <span class="font-normal">({{ $t('setting.hasNewVersion') }})</span>
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

        <Upgrade ref="upgradeRef" @search="search" />
    </div>
</template>

<script setup lang="ts">
import { getSettingInfo, loadUpgradeInfo } from '@/api/modules/setting';
import Upgrade from '@/components/system-upgrade/upgrade/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const upgradeRef = ref();

const version = ref<string>('');
const isMasterProductPro = ref();
const loading = ref(false);
const upgradeInfo = ref();
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
                upgradeRef.value.acceptParams({ upgradeInfo: upgradeInfo.value, upgradeVersion: upgradeVersion.value });
            } else {
                MsgSuccess(i18n.global.t('setting.noUpgrade'));
                return;
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    isMasterProductPro.value = globalStore.isMasterProductPro;
    search();
});
</script>

<style lang="scss" scoped>
.version {
    font-size: 14px;
    color: var(--dark-gold-base-color);
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
        color: var(--dark-gold-base-color);
        margin: 13px, 0;
        padding: 0;
        font-size: 16px;
    }
}
</style>
