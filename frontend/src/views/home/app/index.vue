<template>
    <div>
        <CardWithHeader :header="$t('app.app')" class="mt-5" :loading="loading">
            <template #header-r>
                <el-popover placement="left" :width="226" trigger="click">
                    <el-input size="small" v-model="filter" clearable @input="loadOption()" />
                    <el-table :show-header="false" :data="options" max-height="500px">
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
                        <el-button class="h-button-setting" link icon="Setting"></el-button>
                    </template>
                </el-popover>
            </template>
            <template #body>
                <el-scrollbar height="545px" class="moz-height">
                    <div class="h-app-card" v-for="(app, index) in apps" :key="index">
                        <el-row :gutter="5">
                            <el-col :span="5">
                                <div>
                                    <el-badge
                                        badge-style="background-color: transparent; font-size: 10px; border: none"
                                        v-if="app.isRecommend"
                                        :offset="[-60, 0]"
                                    >
                                        <template #content>
                                            <svg-icon iconName="p-tuijian"></svg-icon>
                                        </template>
                                        <el-avatar
                                            shape="square"
                                            :size="55"
                                            :src="'data:image/png;base64,' + app.icon"
                                        />
                                    </el-badge>
                                    <el-avatar
                                        v-else
                                        shape="square"
                                        :size="55"
                                        :src="'data:image/png;base64,' + app.icon"
                                    />
                                </div>
                            </el-col>
                            <el-col :span="16">
                                <div class="h-app-content" v-if="!app.currentRow">
                                    <div>
                                        <span class="h-app-title">{{ app.name }}</span>
                                    </div>
                                    <div class="h-app-desc">
                                        <span>
                                            {{ app.description }}
                                        </span>
                                    </div>
                                </div>
                                <div class="h-app-content" v-else>
                                    <div>
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
                                    <div class="h-app-margin">
                                        <el-button plain size="small" link class="h-app-desc">
                                            {{ $t('app.version') + ': ' + app.currentRow.version }}
                                        </el-button>
                                    </div>
                                    <div class="h-app-margin">
                                        <el-button
                                            size="small"
                                            type="primary"
                                            link
                                            v-if="app.currentRow.status !== 'Running'"
                                            @click="onOperate('start', app.currentRow)"
                                        >
                                            {{ $t('commons.button.start') }}
                                        </el-button>
                                        <el-button
                                            :style="mobile ? 'margin-left: -1px' : ''"
                                            size="small"
                                            type="primary"
                                            link
                                            v-else
                                            @click="onOperate('stop', app.currentRow)"
                                        >
                                            {{ $t('commons.button.stop') }}
                                        </el-button>
                                        <el-button
                                            :style="mobile ? 'margin-left: -1px' : ''"
                                            size="small"
                                            type="primary"
                                            link
                                            @click="onOperate('restart', app.currentRow)"
                                        >
                                            {{ $t('commons.button.restart') }}
                                        </el-button>
                                        <el-button
                                            :style="mobile ? 'margin-left: -1px' : ''"
                                            size="small"
                                            type="primary"
                                            link
                                            @click="toFolder(app.currentRow.path)"
                                        >
                                            {{ $t('home.dir') }}
                                        </el-button>
                                        <el-popover
                                            placement="left"
                                            trigger="hover"
                                            v-if="app.currentRow.appType == 'website'"
                                            :width="260"
                                        >
                                            <template #reference>
                                                <el-button
                                                    link
                                                    size="small"
                                                    type="primary"
                                                    :style="mobile ? 'margin-left: -1px' : ''"
                                                >
                                                    {{ $t('app.toLink') }}
                                                </el-button>
                                            </template>
                                            <span v-if="defaultLink == '' && app.currentRow.webUI == ''">
                                                {{ $t('app.webUIConfig') }}
                                            </span>
                                            <div v-else>
                                                <div>
                                                    <el-button
                                                        v-if="defaultLink != ''"
                                                        type="primary"
                                                        link
                                                        @click="toLink(defaultLink + ':' + app.currentRow.httpPort)"
                                                    >
                                                        {{ defaultLink + ':' + app.currentRow.httpPort }}
                                                    </el-button>
                                                </div>
                                                <div>
                                                    <el-button
                                                        v-if="app.currentRow.webUI != ''"
                                                        type="primary"
                                                        link
                                                        @click="toLink(app.currentRow.webUI)"
                                                    >
                                                        {{ app.currentRow.webUI }}
                                                    </el-button>
                                                </div>
                                            </div>
                                        </el-popover>
                                    </div>
                                </div>
                            </el-col>
                            <el-col :span="1">
                                <el-button
                                    class="h-app-button"
                                    type="primary"
                                    plain
                                    round
                                    size="small"
                                    :disabled="app.limit == 1 && app.detail && app.detail.length !== 0"
                                    @click="goInstall(app.key, app.appType)"
                                >
                                    {{ $t('commons.button.install') }}
                                </el-button>
                            </el-col>
                        </el-row>
                        <div class="h-app-divider" />
                    </div>
                </el-scrollbar>
            </template>
        </CardWithHeader>
    </div>
</template>

<script lang="ts" setup>
import { getAppStoreConfig, installedOp } from '@/api/modules/app';
import { changeLauncherStatus, loadAppLauncher, loadAppLauncherOption } from '@/api/modules/dashboard';
import i18n from '@/lang';
import { GlobalStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { toFolder } from '@/global/business';

const router = useRouter();
const globalStore = GlobalStore();

let loading = ref(false);
let apps = ref([]);
const options = ref([]);
const filter = ref();
const mobile = computed(() => {
    return globalStore.isMobile();
});
const defaultLink = ref('');

const acceptParams = (): void => {
    search();
    loadOption();
    getConfig();
};

const goInstall = (key: string, type: string) => {
    switch (type) {
        case 'php':
        case 'node':
        case 'java':
        case 'go':
        case 'python':
        case 'dotnet':
            router.push({ path: '/websites/runtimes/' + type });
            break;
        default:
            router.push({ name: 'AppAll', query: { install: key } });
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

const toLink = (link: string) => {
    window.open(link, '_blank');
};

const getConfig = async () => {
    try {
        const res = await getAppStoreConfig();
        if (res.data.defaultDomain != '') {
            defaultLink.value = res.data.defaultDomain;
        }
    } catch (error) {}
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

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.h-app-card {
    cursor: pointer;
    padding: 10px 15px;
    margin-right: 10px;
    line-height: 18px;

    .h-app-content {
        padding-left: 15px;
        .h-app-title {
            font-weight: 500;
            font-size: 15px;
            color: var(--panel-text-color);
        }

        .h-app-desc {
            span {
                font-weight: 400;
                font-size: 12px;
                color: var(--el-text-color-regular);
            }
        }
    }
    .h-app-button {
        margin-top: 10px;
    }
    &:hover {
        background-color: rgba(0, 94, 235, 0.03);
    }
}

.h-app-divider {
    margin-top: 10px;
    border: 0;
    border-top: var(--panel-border);
}

.h-app-desc {
    font-weight: 400;
    font-size: 12px;
    color: var(--el-text-color-regular);
}

.h-button-setting {
    float: right;
    margin-left: 5px;
}

.h-app-dropdown {
    font-weight: 600;
    font-size: 16px;
    color: var(--panel-text-color);
}

.h-app-margin {
    margin-top: 2px;
}

.h-app-option {
    font-weight: 500;
    font-size: 14px;
    line-height: 20px;
    color: var(--el-text-color-regular);
}

/* FOR MOZILLA */
@-moz-document url-prefix() {
    .moz-height {
        height: 524px;
    }
}
</style>
