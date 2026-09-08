<template>
    <div class="install-card">
        <el-card class="e-card">
            <el-row :gutter="10" class="install-card-row">
                <el-col class="install-card-icon-col" :xs="3" :sm="3" :md="3" :lg="4" :xl="3">
                    <AppIcon
                        @open-detail="$emit('openDetail')"
                        :appKey="installed.appKey"
                        :currentNode="currentNode"
                    ></AppIcon>
                </el-col>
                <el-col class="install-card-detail-col" :xs="21" :sm="21" :md="21" :lg="20" :xl="21">
                    <div class="a-detail">
                        <AppHeader
                            :installed="installed"
                            :mode="mode"
                            :defaultLink="defaultLink"
                            :sortMode="sortMode"
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
    sortMode?: boolean;
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

@media only screen and (max-width: 1023px) {
    .install-card-row {
        display: grid;
        grid-template-columns: 87px minmax(0, 1fr);
    }

    .install-card-icon-col,
    .install-card-detail-col {
        width: auto;
        max-width: none;
        min-width: 0;
        flex: initial;
    }
}

@media only screen and (max-width: 767px) {
    .install-card-row {
        grid-template-columns: 74px minmax(0, 1fr);
    }

    .install-card-icon-col {
        :deep(.el-avatar) {
            max-width: 64px;
            max-height: 64px;
        }
    }

    .install-card-row .install-card-detail-col .a-detail {
        min-width: 0;

        :deep(.d-name .d-name-row) {
            min-width: 0;
            align-items: flex-start;
            flex-direction: column;
        }

        :deep(.d-name .d-name-row .name-actions) {
            width: 100%;
            min-width: 0;
            flex: 0 1 auto;
            flex-wrap: wrap;
            overflow: visible;
        }

        :deep(.d-name .d-name-row .name-wrap) {
            min-width: min(100px, 100%);
            max-width: calc(100% - 56px);
            flex: 1 1 100px;
        }

        :deep(.d-name .d-name-row .operate-actions) {
            width: 100%;
            min-width: 0;
            flex-shrink: 1;
            justify-content: flex-start;

            .el-button + .el-button {
                margin-left: 0;
            }

            .h-button {
                margin: 0 !important;
            }
        }

        :deep(.d-description),
        :deep(.d-button) {
            width: 100%;
            max-width: 100%;
            min-width: 0;
        }

        :deep(.d-description .el-button) {
            max-width: 100%;
            height: auto;
            min-height: 24px;
            white-space: normal;

            > span {
                min-width: 0;
                overflow-wrap: anywhere;
            }
        }
    }
}
</style>
