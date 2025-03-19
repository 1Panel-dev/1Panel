<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.about')" :divider="true">
            <template #main>
                <div style="text-align: center; margin-top: 20px">
                    <div style="justify-self: center" class="logo">
                        <img
                            v-if="globalStore.themeConfig.logo"
                            style="width: 80px"
                            :src="`/api/v2/images/logo?t=${Date.now()}`"
                        />
                        <PrimaryLogo v-else />
                    </div>
                    <h3 class="description">{{ globalStore.themeConfig.title || $t('setting.description') }}</h3>
                    <div class="flex justify-center">
                        <SystemUpgrade class="upgrade" />
                    </div>
                    <div class="flex w-full justify-center my-5 flex-wrap md:flex-row gap-4">
                        <el-link @click="toDoc" class="system-link">
                            <el-icon><Document /></el-icon>
                            <span>{{ $t('setting.doc2') }}</span>
                        </el-link>
                        <el-link @click="toGithub" class="system-link">
                            <svg-icon iconName="p-huaban88"></svg-icon>
                            <span>{{ $t('setting.project') }}</span>
                        </el-link>
                        <el-link @click="toIssue" class="system-link">
                            <svg-icon iconName="p-bug"></svg-icon>
                            <span>{{ $t('setting.issue') }}</span>
                        </el-link>
                        <el-link @click="toGithubStar" class="system-link">
                            <svg-icon iconName="p-star"></svg-icon>
                            <span>{{ $t('setting.star') }}</span>
                        </el-link>
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { getSystemAvailable } from '@/api/modules/setting';
import { onMounted, ref } from 'vue';
import SystemUpgrade from '@/components/system-upgrade/index.vue';
import { GlobalStore } from '@/store';
import PrimaryLogo from '@/assets/images/1panel-logo.svg?component';
import { storeToRefs } from 'pinia';
const globalStore = GlobalStore();
const { docsUrl } = storeToRefs(globalStore);
const loading = ref();

const toDoc = () => {
    window.open(docsUrl.value, '_blank', 'noopener,noreferrer');
};
const toGithub = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank', 'noopener,noreferrer');
};
const toIssue = () => {
    window.open('https://github.com/1Panel-dev/1Panel/issues', '_blank', 'noopener,noreferrer');
};
const toGithubStar = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank', 'noopener,noreferrer');
};

onMounted(() => {
    getSystemAvailable();
});
</script>

<style lang="scss" scoped>
.system-link {
    margin-left: 15px;

    .svg-icon {
        font-size: 7px;
    }
    span {
        line-height: 20px;
        font-weight: 400;
    }
}
.description {
    color: var(--el-text-color-regular);
}
.logo {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 55px;
    img {
        object-fit: contain;
        width: 95%;
        height: 45px;
    }
}
.upgrade {
    all: initial;
}
</style>
