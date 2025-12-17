<template>
    <div class="d-name">
        <div class="flex items-center justify-between">
            <div class="min-w-50 flex items-center justify-start gap-1">
                <el-button link type="info">
                    <el-tooltip :content="installed.name" placement="top">
                        <span class="name">{{ installed.name }}</span>
                    </el-tooltip>
                </el-button>
                <span class="status">
                    <Status :key="installed.status" :status="installed.status"></Status>
                </span>
                <span class="msg">
                    <el-popover
                        v-if="isAppErr(installed)"
                        placement="bottom"
                        :width="400"
                        trigger="hover"
                        :content="installed.message"
                        :popper-options="options"
                    >
                        <template #reference>
                            <el-button link type="danger">
                                <el-icon><Warning /></el-icon>
                            </el-button>
                        </template>
                        <div class="app-error">
                            {{ installed.message }}
                        </div>
                    </el-popover>
                </span>
                <span class="ml-1">
                    <el-tooltip effect="dark" :content="$t('app.toFolder')" placement="top">
                        <el-button type="primary" link @click="$emit('toFolder')" icon="FolderOpened"></el-button>
                    </el-tooltip>
                </span>
                <span class="ml-1">
                    <el-tooltip
                        v-if="mode !== 'upgrade'"
                        effect="dark"
                        :content="$t('commons.button.log')"
                        placement="top"
                    >
                        <el-button
                            type="primary"
                            link
                            @click="$emit('openLog')"
                            :disabled="installed.status === 'DownloadErr'"
                        >
                            <el-icon><Tickets /></el-icon>
                        </el-button>
                    </el-tooltip>
                </span>
                <span class="ml-1">
                    <el-tooltip v-if="mode !== 'upgrade'" effect="dark" :content="$t('menu.terminal')" placement="top">
                        <el-button
                            type="primary"
                            link
                            @click="$emit('openTerminal')"
                            :disabled="installed.status !== 'Running'"
                        >
                            <el-icon>
                                <SvgIcon iconName="p-terminal2" />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                </span>
                <span class="ml-1">
                    <el-tooltip v-if="mode !== 'upgrade'" effect="dark" :content="$t('menu.container')" placement="top">
                        <el-button type="primary" link @click="$emit('toContainer')">
                            <el-icon>
                                <SvgIcon iconName="p-docker" />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                </span>
                <span class="ml-1" v-if="mode === 'installed'">
                    <el-tooltip
                        effect="dark"
                        :content="$t('website.cancelFavorite')"
                        placement="top-start"
                        v-if="installed.favorite"
                    >
                        <el-button
                            link
                            size="large"
                            icon="StarFilled"
                            type="warning"
                            @click="$emit('favoriteInstall')"
                        ></el-button>
                    </el-tooltip>
                    <el-tooltip effect="dark" :content="$t('website.favorite')" placement="top-start" v-else>
                        <el-button link icon="Star" type="info" @click="$emit('favoriteInstall')"></el-button>
                    </el-tooltip>
                </span>
            </div>
            <div class="flex flex-wrap items-center justify-end gap-1">
                <el-button
                    class="h-button"
                    plain
                    round
                    size="small"
                    @click="$emit('openUploads')"
                    v-if="mode === 'installed'"
                >
                    {{ $t('database.loadBackup') }}
                </el-button>
                <el-button
                    class="h-button"
                    plain
                    round
                    size="small"
                    @click="$emit('openBackups')"
                    v-if="mode === 'installed'"
                >
                    {{ $t('commons.button.backup') }}
                </el-button>
                <el-button
                    class="h-button"
                    plain
                    round
                    size="small"
                    :disabled="installed.status === 'Upgrading'"
                    @click="$emit('ignoreApp')"
                    v-if="mode === 'upgrade'"
                >
                    {{ $t('commons.button.ignore') }}
                </el-button>
                <el-button
                    class="h-button"
                    plain
                    round
                    size="small"
                    :disabled="
                        (installed.status !== 'Running' && installed.status !== 'UpgradeErr') ||
                        installed.appStatus === 'TakeDown'
                    "
                    @click="$emit('openOperate')"
                    v-if="mode === 'upgrade'"
                >
                    {{ $t('commons.button.upgrade') }}
                </el-button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';

interface Props {
    installed: App.AppInstalled;
    mode: string;
}
defineProps<Props>();

defineEmits([
    'toFolder',
    'openLog',
    'openTerminal',
    'toContainer',
    'openBackups',
    'openOperate',
    'ignoreApp',
    'openUploads',
    'favoriteInstall',
]);

const options = {
    modifiers: [
        {
            name: 'flip',
            options: {
                padding: 5,
                fallbackPlacements: ['bottom-start', 'top-start', 'right', 'left'],
            },
        },
    ],
};

const isAppErr = (row: any) => {
    return row.status.includes('Err') || row.status.includes('Error') || row.status.includes('UnHealthy');
};
</script>
<style scoped lang="scss">
@use '@/views/app-store/index.scss';

.svg-icon {
    width: 100%;
    height: 100%;
    padding: 0;
}
</style>
