<template>
    <div class="app">
        <el-card>
            <div class="app-wrapper" @click="openDetail(app.key)">
                <div class="app-image">
                    <el-avatar shape="square" :size="60" :src="'data:image/png;base64,' + app.icon" />
                </div>
                <div class="app-content">
                    <div class="content-top">
                        <el-space wrap :size="1">
                            <span class="app-title">{{ app.name }}</span>
                            <el-tag type="success" v-if="app.installed" round size="small" class="!ml-2">
                                {{ $t('app.allReadyInstalled') }}
                            </el-tag>
                        </el-space>
                    </div>
                    <div class="content-middle">
                        <span class="app-description">
                            {{ app.description }}
                        </span>
                    </div>
                    <div class="content-bottom">
                        <div class="app-tags">
                            <el-tag v-for="(tag, ind) in app.tags" :key="ind" type="info">
                                <span>
                                    {{ tag.name }}
                                </span>
                            </el-tag>
                            <el-tag v-if="app.status === 'TakeDown'" class="p-mr-5">
                                <span class="text-red-500">{{ $t('app.takeDown') }}</span>
                            </el-tag>
                        </div>
                        <el-button
                            type="primary"
                            size="small"
                            plain
                            round
                            :disabled="(app.installed && app.limit == 1) || app.status === 'TakeDown'"
                            @click.stop="openInstall(app)"
                        >
                            {{ $t('commons.button.install') }}
                        </el-button>
                    </div>
                </div>
            </div>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
defineProps({
    app: {
        type: Object,
        default: () => ({}),
    },
});

const em = defineEmits(['openDetail', 'openInstall']);

const openDetail = (key: string) => {
    em('openDetail', key);
};

const openInstall = (app: any) => {
    em('openInstall', app);
};
</script>

<style lang="scss" scoped>
.app {
    margin: 10px;
    .el-card {
        padding: 0 !important;
        border: var(--panel-border) !important;
        &:hover {
            border: 1px solid var(--el-color-primary) !important;
        }
    }
    .el-card__body {
        padding: 8px 8px 2px 8px !important;
    }
    .app-wrapper {
        display: flex;
        height: 100%;
        cursor: pointer;
    }
    .app-image {
        flex: 0 0 100px;
        display: flex;
        justify-content: center;
        margin-top: 14px;
        transition: transform 0.1s;
    }

    &:hover .app-image {
        transform: scale(1.2);
    }

    .el-avatar {
        width: 65px !important;
        height: 65px !important;
        max-width: 65px;
        max-height: 65px;
        object-fit: cover;
    }
    .app-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        padding: 10px;
    }
    .content-top,
    .content-bottom {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .content-middle {
        flex: 1;
        margin: 10px 0;
        overflow: hidden; /* 防止内容溢出 */
    }
    .app-name {
        margin: 0;
        line-height: 1.5;
        font-weight: 500;
        font-size: 16px;
        color: var(--el-text-color-regular);
    }
    .app-description {
        margin: 0;
        overflow: hidden;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        text-overflow: ellipsis;
        font-size: 14px;
        color: var(--el-text-color-regular);

        line-height: 1.2;
        height: calc(1.2em * 2);
        min-height: calc(1.2em * 2);
    }
    .app-tags {
        display: flex;
        gap: 5px;
    }
}
</style>
