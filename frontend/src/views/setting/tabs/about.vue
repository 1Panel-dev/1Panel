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
            <h1>1Panel</h1>
            <h3>{{ $t('setting.description') }}</h3>
            <h3>v0.0.1</h3>
            <div style="margin-top: 10px">
                <svg-icon style="font-size: 9px" iconName="p-huaban88"></svg-icon>
                <el-link @click="toGithub">
                    <span>{{ $t('setting.project') }}</span>
                </el-link>
                <svg-icon style="font-size: 9px; margin-left: 15px" iconName="p-bug"></svg-icon>
                <el-link @click="toIssue">
                    <span>{{ $t('setting.issue') }}</span>
                </el-link>
                <svg-icon style="font-size: 9px; margin-left: 15px" iconName="p-taolun"></svg-icon>
                <el-link @click="toTalk">
                    <span>{{ $t('setting.chat') }}</span>
                </el-link>
                <svg-icon style="font-size: 9px; margin-left: 15px" iconName="p-star"></svg-icon>
                <el-link @click="toGithubStar">
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
