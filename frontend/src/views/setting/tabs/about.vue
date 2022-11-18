<template>
    <el-card style="margin-top: 20px">
        <template #header>
            <div class="card-header">
                <span>{{ $t('setting.about') }}</span>
            </div>
        </template>
        <div style="text-align: center">
            <div style="justify-self: center">
                <img style="width: 80px" src="@/assets/images/ko_image.png" />
            </div>
            <h3>{{ $t('setting.description') }}</h3>
            <h3>v1.0.0</h3>
            <div style="margin-top: 10px">
                <el-link @click="toGithub">
                    <svg-icon style="font-size: 9px; margin-bottom: 8px" iconName="p-huaban88"></svg-icon>
                    <span style="line-height: 20px">{{ $t('setting.project') }}</span>
                </el-link>
                <el-link @click="toIssue" style="margin-left: 15px">
                    <svg-icon style="font-size: 9px; margin-bottom: 8px" iconName="p-bug"></svg-icon>
                    <span>{{ $t('setting.issue') }}</span>
                </el-link>
                <el-link @click="toTalk" style="margin-left: 15px">
                    <svg-icon style="font-size: 9px; margin-bottom: 8px" iconName="p-taolun"></svg-icon>
                    <span>{{ $t('setting.chat') }}</span>
                </el-link>
                <el-link @click="toGithubStar" style="margin-left: 15px">
                    <svg-icon style="font-size: 9px; margin-bottom: 6px" iconName="p-star"></svg-icon>
                    <span>{{ $t('setting.star') }}</span>
                </el-link>
            </div>
        </div>
    </el-card>
</template>

<script lang="ts" setup>
const toGithub = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank');
};
const toIssue = () => {
    window.open('https://github.com/1Panel-dev/1Panel/issues', '_blank');
};
const toTalk = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank');
};
const toGithubStar = () => {
    window.open('https://github.com/1Panel-dev/1Panel', '_blank');
};
</script>
