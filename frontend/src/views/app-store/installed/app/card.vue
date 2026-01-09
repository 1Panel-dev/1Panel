<template>
    <div class="install-card">
        <el-card class="e-card">
            <el-row :gutter="10">
                <el-col :xs="3" :sm="3" :md="3" :lg="4" :xl="3">
                    <AppIcon
                        @open-detail="$emit('openDetail')"
                        :appKey="installed.appKey"
                        :currentNode="currentNode"
                    ></AppIcon>
                </el-col>
                <el-col :xs="21" :sm="21" :md="21" :lg="20" :xl="21">
                    <div class="a-detail">
                        <AppHeader
                            :installed="installed"
                            :mode="mode"
                            :defaultLink="defaultLink"
                            @open-backups="$emit('openBackups')"
                            @open-log="$emit('openLog')"
                            @open-terminal="$emit('openTerminal')"
                            @open-operate="$emit('openOperate')"
                            @favorite-install="$emit('favoriteInstall')"
                            @to-folder="$emit('toFolder')"
                            @open-uploads="$emit('openUploads')"
                            @ignore-app="$emit('ignoreApp')"
                            @to-container="$emit('toContainer')"
                        ></AppHeader>
                        <AppInfo
                            :installed="installed"
                            :defaultLink="defaultLink"
                            @jump-to-path="$emit('jumpToPath', '/settings/panel')"
                        ></AppInfo>
                        <div class="app-divider" />
                        <slot name="buttons"></slot>
                    </div>
                </el-col>
            </el-row>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import AppIcon from '@/views/app-store/installed/app/icon.vue';
import AppHeader from '@/views/app-store/installed/app/header.vue';
import AppInfo from '@/views/app-store/installed/app/info.vue';

import { App } from '@/api/interface/app';

interface Props {
    installed: App.AppInstalled;
    mode: string;
    defaultLink: string;
    currentNode: string;
}
defineProps<Props>();

defineEmits([
    'toFolder',
    'openUploads',
    'openDetail',
    'openBackups',
    'openLog',
    'openTerminal',
    'openOperate',
    'favoriteInstall',
    'jumpToPath',
    'ignoreApp',
    'toContainer',
]);
</script>

<style scoped lang="scss">
@use '@/views/app-store/index.scss';
</style>
