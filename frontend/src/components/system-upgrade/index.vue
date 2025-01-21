<template>
    <div>
        <div class="flex w-full flex-col gap-2 md:flex-row items-center">
            <div class="flex flex-wrap items-center" v-if="props.footer">
                <el-link type="primary" :underline="false" @click="toForum">
                    <span class="font-normal">{{ $t('setting.forum') }}</span>
                </el-link>
                <el-divider direction="vertical" />
                <el-link type="primary" :underline="false" @click="toDoc">
                    <span class="font-normal">{{ $t('setting.doc2') }}</span>
                </el-link>
                <el-divider direction="vertical" />
                <el-link type="primary" :underline="false" @click="toGithub">
                    <span class="font-normal">{{ $t('setting.project') }}</span>
                </el-link>
                <el-divider direction="vertical" />
            </div>
            <div class="flex flex-wrap items-center">
                <el-link :underline="false" type="primary" @click="toLxware">
                    {{ $t(!isMasterProductPro ? 'license.community' : 'license.pro') }}
                </el-link>
                <el-link :underline="false" class="version" type="primary" @click="copyText(version)">
                    {{ version }}
                </el-link>
                <el-badge is-dot class="-mt-0.5" v-if="version !== 'Waiting' && globalStore.hasNewVersion">
                    <el-link :underline="false" type="primary" @click="onLoadUpgradeInfo">
                        （{{ $t('setting.hasNewVersion') }}）
                    </el-link>
                </el-badge>
                <el-link
                    v-if="version !== 'Waiting' && !globalStore.hasNewVersion"
                    type="primary"
                    :underline="false"
                    @click="onLoadUpgradeInfo"
                >
                    （{{ $t('setting.upgradeCheck') }}）
                </el-link>
                <el-tag v-if="version === 'Waiting'" round style="margin-left: 10px">
                    {{ $t('setting.upgrading') }}
                </el-tag>
            </div>
        </div>

        <Upgrade ref="upgradeRef" @search="search" />
    </div>
</template>

<script setup lang="ts">
import { getSettingInfo, loadUpgradeInfo } from '@/api/modules/setting';
import Upgrade from '@/components/system-upgrade/upgrade/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { copyText } from '@/utils/util';
import { onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { docsUrl } = storeToRefs(globalStore);
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

const toLxware = () => {
    if (!globalStore.isIntl) {
        window.open('https://www.lxware.cn/1panel' + '', '_blank', 'noopener,noreferrer');
    } else {
        window.open('https://1panel.hk/pricing' + '', '_blank', 'noopener,noreferrer');
    }
};

const toDoc = () => {
    window.open(docsUrl.value, '_blank', 'noopener,noreferrer');
};

const toForum = () => {
    let url = globalStore.isIntl
        ? 'https://github.com/1Panel-dev/1Panel/discussions'
        : 'https://bbs.fit2cloud.com/c/1p/7';
    window.open(url, '_blank');
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
    color: var(--panel-color-primary-light-4);
    text-decoration: none;
    letter-spacing: 0.5px;
}
</style>
