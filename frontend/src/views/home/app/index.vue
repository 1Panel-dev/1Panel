<template>
    <div>
        <CardWithHeader :header="$t('app.app')" class="card-interval" v-loading="loading">
            <template #header-r>
                <el-popover placement="left" :width="226" trigger="click">
                    <el-input size="small" v-model="filter" clearable @input="loadOption()" />
                    <el-table :show-header="false" :data="options" max-height="150px">
                        <el-table-column prop="key" width="120" show-overflow-tooltip />
                        <el-table-column prop="name">
                            <template #default="{ row }">
                                <el-switch
                                    @change="onChangeStatus(row)"
                                    class="float-right"
                                    size="small"
                                    v-model="row.isShow"
                                />
                            </template>
                        </el-table-column>
                    </el-table>
                    <template #reference>
                        <el-button
                            class="h-button-setting"
                            :disabled="!globalStore.isAdminOrNodeAdmin"
                            link
                            icon="Setting"
                        ></el-button>
                    </template>
                </el-popover>
            </template>
            <template #body>
                <el-scrollbar height="536px" class="moz-height">
                    <div class="h-app-card" v-for="(app, index) in apps" :key="index">
                        <el-row :gutter="5">
                            <el-col :span="5" class="h-app-media-col">
                                <div class="h-app-media">
                                    <el-avatar shape="square" :size="55" :src="getAppIconSrc(app)" />
                                </div>
                            </el-col>
                            <el-col :span="16" class="h-app-main-col">
                                <div class="h-app-content" v-if="!app.currentRow">
                                    <div class="h-app-heading">
                                        <span class="h-app-title">{{ app.name }}</span>
                                        <svg-icon class="svg-icon" iconName="p-huobao1"></svg-icon>
                                    </div>
                                    <div class="h-app-desc">
                                        <span>
                                            {{ app.description }}
                                        </span>
                                    </div>
                                </div>
                                <div class="h-app-content" v-else>
                                    <div class="h-app-heading">
                                        <el-dropdown trigger="hover">
                                            <el-button type="primary" plain size="small" link class="h-app-dropdown">
                                                {{ app.currentRow.name }}
                                                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                                            </el-button>
                                            <template #dropdown>
                                                <el-dropdown-menu>
                                                    <el-dropdown-item
                                                        v-for="(detailItem, index2) in app.detail"
                                                        :key="index2"
                                                        @click="app.currentRow = detailItem"
                                                    >
                                                        {{ detailItem.name + ' - ' + detailItem.version }}
                                                    </el-dropdown-item>
                                                </el-dropdown-menu>
                                            </template>
                                        </el-dropdown>
                                    </div>
                                    <div class="h-app-meta">
                                        <el-button plain size="small" link class="h-app-desc">
                                            {{ $t('app.version') + ': ' + app.currentRow.version }}
                                        </el-button>
                                    </div>
                                    <div class="h-app-actions">
                                        <el-button
                                            size="small"
                                            type="primary"
                                            link
                                            v-if="app.currentRow.status !== 'Running'"
                                            :disabled="!hasManagePermission"
                                            @click="onOperate('start', app.currentRow)"
                                        >
                                            {{ $t('commons.button.start') }}
                                        </el-button>
                                        <el-button
                                            :style="isMobile ? 'margin-left: -1px' : ''"
                                            size="small"
                                            type="primary"
                                            link
                                            v-else
                                            :disabled="!hasManagePermission"
                                            @click="onOperate('stop', app.currentRow)"
                                        >
                                            {{ $t('commons.button.stop') }}
                                        </el-button>
                                        <el-button
                                            :style="isMobile ? 'margin-left: -1px' : ''"
                                            size="small"
                                            type="primary"
                                            link
                                            :disabled="!hasManagePermission"
                                            @click="onOperate('restart', app.currentRow)"
                                        >
                                            {{ $t('commons.button.restart') }}
                                        </el-button>
                                        <el-button
                                            :style="isMobile ? 'margin-left: -1px' : ''"
                                            size="small"
                                            type="primary"
                                            link
                                            :disabled="!hasManagePermission"
                                            @click="routerToName('AppInstalled')"
                                        >
                                            {{ $t('tabs.more') }}
                                        </el-button>
                                    </div>
                                </div>
                            </el-col>
                            <el-col :span="3" class="h-app-side-col">
                                <el-button
                                    class="h-app-button"
                                    type="primary"
                                    plain
                                    round
                                    size="small"
                                    :disabled="
                                        (app.limit == 1 && app.detail && app.detail.length !== 0) ||
                                        !hasManagePermission
                                    "
                                    @click="goInstall(app.key, app.appType)"
                                >
                                    {{ $t('commons.button.install') }}
                                </el-button>
                            </el-col>
                        </el-row>
                    </div>
                </el-scrollbar>
            </template>
        </CardWithHeader>
    </div>
</template>

<script lang="ts" setup>
import { getAppIconUrl, installedOp } from '@/api/modules/app';
import { changeLauncherStatus, loadAppLauncher, loadAppLauncherOption } from '@/api/modules/dashboard';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ref } from 'vue';
import { jumpToInstall } from '@/utils/app';
import { routerToName, routerToNameWithQuery } from '@/utils/router';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile } = useGlobalStore();
let loading = ref(false);
let apps = ref([]);
const options = ref([]);
const filter = ref();

const acceptParams = (): void => {
    search();
    loadOption();
};

const goInstall = (key: string, type: string) => {
    if (!jumpToInstall(type, key)) {
        routerToNameWithQuery('AppAll', { install: key });
    }
};

const search = async () => {
    loading.value = true;
    await loadAppLauncher()
        .then((res) => {
            loading.value = false;
            apps.value = res.data;
            for (const item of apps.value) {
                if (item.detail && item.detail.length !== 0) {
                    item.currentRow = item.detail[0];
                }
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const onChangeStatus = async (row: any) => {
    loading.value = true;
    await changeLauncherStatus(row.key, row.isShow ? 'Enable' : 'Disable')
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onOperate = async (operation: string, row: any) => {
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.button.' + operation)]),
        i18n.global.t('commons.button.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        loading.value = true;
        let params = {
            installId: row.installID,
            operate: operation,
            detailId: row.detailID,
        };
        await installedOp(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const loadOption = async () => {
    const res = await loadAppLauncherOption(filter.value || '');
    options.value = res.data || [];
};

const getAppIconSrc = (app: any) => {
    const icon = app?.icon || '';
    if (icon.startsWith('app_') || icon === '') {
        return getAppIconUrl(app.key);
    }
    return `data:image/png;base64,${icon}`;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.h-app-card {
    padding: 8px 10px;
    margin: 0 4px 3px 0;
    min-height: 66px;
    line-height: 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    transition: border-color 0.18s ease;

    .h-app-content {
        padding-left: 4px;

        .h-app-title {
            font-weight: 600;
            font-size: 14px;
            color: var(--panel-text-color);
            letter-spacing: 0.01em;
        }

        .h-app-desc {
            margin-top: 3px;
            line-height: 1.4;

            span {
                font-weight: 400;
                font-size: 11px;
                color: var(--el-text-color-regular);
                display: -webkit-box;
                overflow: hidden;
                -webkit-box-orient: vertical;
                -webkit-line-clamp: 1;
            }
        }
    }

    .h-app-button {
        padding: 1px 10px;
        line-height: 1;
        border-radius: 999px;
    }

    &:hover {
        border-color: var(--el-color-primary);
    }
}

.h-app-media-col,
.h-app-side-col {
    display: flex;
    align-items: flex-start;
}

.h-app-side-col {
    align-items: center;
    min-height: 58px;
    justify-content: flex-end;
}

.h-app-main-col {
    display: flex;
    align-items: center;
    padding-right: 6px;
}

.h-app-media {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 58px;
    height: 58px;
    border-radius: 16px;

    :deep(.el-avatar) {
        width: 50px !important;
        height: 50px !important;
        border-radius: 14px;
    }
}

.h-app-heading {
    display: flex;
    align-items: center;
    gap: 4px;
    min-height: 20px;
    line-height: 1.2;
}

.h-app-meta {
    margin-top: 1px;
}

.h-app-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0 10px;
    margin-top: 2px;
    line-height: 1.2;
}

.h-app-desc {
    font-weight: 400;
    font-size: 11px;
    color: var(--el-text-color-regular);
}

.svg-icon {
    font-size: 5px !important;
    margin-left: 5px;
    margin-bottom: 2px;
}

.h-button-setting {
    float: right;
    margin-left: 5px;
}

.h-app-dropdown {
    font-weight: 600;
    font-size: 14px;
    color: var(--panel-text-color);
    min-height: 18px;
}

:deep(.h-app-actions .el-button) {
    min-height: 16px;
    margin-left: 0 !important;
    padding: 0;
}

@media only screen and (max-width: 768px) {
    .h-app-card {
        padding: 12px;
        margin-right: 0;
        border-radius: 14px;
    }

    .h-app-card .h-app-content {
        padding-left: 0;
    }

    .h-app-media {
        width: 56px;
        height: 56px;
        border-radius: 16px;
    }

    .h-app-heading {
        min-height: 22px;
    }

    .h-app-dropdown,
    .h-app-card .h-app-content .h-app-title {
        font-size: 14px;
    }

    .h-app-actions {
        gap: 2px 8px;
    }
}

/* FOR MOZILLA */
@-moz-document url-prefix() {
    .moz-height {
        height: 524px;
    }
}
</style>
