<template>
    <div>
        <div class="flex w-full flex-col gap-2 md:flex-row items-center">
            <div class="flex flex-wrap gap-y-2 items-center">
                <span v-if="props.footer">
                    <el-link type="primary" underline="never" @click="toForum" v-if="!isFxplay">
                        <span class="font-normal">{{ $t('setting.forum') }}</span>
                    </el-link>
                    <el-divider direction="vertical" v-if="!isFxplay" />
                    <el-link type="primary" underline="never" @click="toDoc">
                        <span class="font-normal">{{ $t('setting.doc2') }}</span>
                    </el-link>
                    <el-divider direction="vertical" v-if="!isFxplay" />
                    <el-link type="primary" underline="never" @click="toGithub" v-if="!isFxplay">
                        <span class="font-normal">{{ $t('setting.project') }}</span>
                    </el-link>
                    <el-divider direction="vertical" />
                </span>
                <div class="flex flex-wrap items-center">
                    <el-link underline="never" type="primary" @click="toLxware">
                        <span v-if="isMasterPro">
                            {{ $t('license.pro') }}
                        </span>
                        <span v-else-if="isOffLine">
                            {{ $t('license.offLine') }}
                        </span>
                        <span v-else>
                            {{ $t('license.community') }}
                        </span>
                    </el-link>
                    <el-link underline="never" class="version" type="primary" @click="getVersionLog()">
                        {{ version }}
                    </el-link>
                    <el-badge is-dot class="-mt-0.5" :hidden="version === 'Waiting' || !globalStore.hasNewVersion">
                        <el-link
                            class="ml-2"
                            underline="never"
                            type="primary"
                            @click="onLoadUpgradeInfo"
                            v-if="!globalStore.isOffLine"
                        >
                            {{ $t('commons.button.update') }}
                        </el-link>
                    </el-badge>
                    <el-tag v-if="version === 'Waiting'" round class="ml-2.5">{{ $t('setting.upgrading') }}</el-tag>
                </div>
            </div>
        </div>

        <Upgrade ref="upgradeRef" @search="search" />
        <Releases ref="releasesRef" />
    </div>
</template>

<script setup lang="ts">
import { getSettingInfo, loadUpgradeInfo } from '@/api/modules/setting';
import Upgrade from '@/components/system-upgrade/upgrade/index.vue';
import Releases from '@/components/system-upgrade/releases/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { docsUrl, isOffLine, isFxplay } = storeToRefs(globalStore);
const upgradeRef = ref();
const releasesRef = ref();
const isMasterPro = computed(() => {
    return globalStore.isMasterPro();
});

const version = ref<string>('');
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

const getVersionLog = () => {
    if (isOffLine.value) {
        return;
    }
    releasesRef.value.acceptParams();
};

const toLxware = () => {
    if (isOffLine.value) {
        to1Panel();
        return;
    }
    if (!globalStore.isIntl) {
        window.open('https://www.lxware.cn/1panel' + '', '_blank', 'noopener,noreferrer');
    } else {
        window.open('https://1panel.hk/pricing' + '', '_blank', 'noopener,noreferrer');
    }
};

const to1Panel = () => {
    window.open('https://1panel.cn', '_blank', 'noopener,noreferrer');
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
                if (upgradeInfo.value.latestVersion) {
                    upgradeVersion.value = upgradeInfo.value.latestVersion;
                } else if (upgradeInfo.value.testVersion) {
                    upgradeVersion.value = upgradeInfo.value.testVersion;
                } else if (upgradeInfo.value.newVersion) {
                    upgradeVersion.value = upgradeInfo.value.newVersion;
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
    search();
});
</script>

<style lang="scss" scoped>
.line-height {
    line-height: 25px;
}
:deep(.el-link__inner) {
    font-weight: 400;
}
.version {
    margin-left: 8px;
    font-size: 14px;
    color: var(--panel-color-primary-light-4);
    text-decoration: none;
    letter-spacing: 0.5px;
    cursor: pointer;
    font-family: auto;
}
</style>
