<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.about')" :divider="true">
            <template #main>
                <div style="text-align: center; margin-top: 20px">
                    <div style="justify-self: center">
                        <img style="width: 80px" src="@/assets/images/1panel-logo-light.png" />
                    </div>
                    <h3>{{ $t('setting.description') }}</h3>
                    <h3>
                        <SystemUpgrade />
                    </h3>
                    <div style="margin-top: 10px">
                        <el-link @click="toDoc">
                            <el-icon><Document /></el-icon>
                            <span>{{ $t('setting.doc') }}</span>
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
import { getSettingInfo, getSystemAvailable } from '@/api/modules/setting';
import { onMounted, ref } from 'vue';
import SystemUpgrade from '@/components/system-upgrade/index.vue';

const version = ref();
const loading = ref();
const search = async () => {
    const res = await getSettingInfo();
    version.value = res.data.systemVersion;
};

const toDoc = () => {
    window.open('https://1panel.cn/docs/', '_blank', 'noopener,noreferrer');
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
    search();
    getSystemAvailable();
});
</script>

<style lang="scss" scoped>
.system-link {
    margin-left: 15px;

    .svg-icon {
        font-size: 7px;
        margin-bottom: 3px;
    }
    span {
        line-height: 20px;
    }
}
</style>
