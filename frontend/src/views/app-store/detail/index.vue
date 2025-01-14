<template>
    <el-drawer
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('app.detail')" :back="handleClose" />
        </template>
        <div class="brief" v-loading="loadingApp">
            <div class="detail flex">
                <div class="w-12 h-12 rounded p-1 shadow-md icon">
                    <img :src="app.icon" alt="App Icon" class="w-full h-full rounded" style="object-fit: contain" />
                </div>
                <div class="ml-4">
                    <div class="name mb-2">
                        <span>{{ app.name }}</span>
                    </div>
                    <div class="description mb-4">
                        <span>
                            {{ app.description }}
                        </span>
                    </div>
                    <br />
                    <div v-if="!loadingDetail" class="mb-2">
                        <el-alert
                            v-if="!appDetail.enable"
                            :title="$t('app.limitHelper')"
                            type="warning"
                            show-icon
                            :closable="false"
                        />
                    </div>
                    <el-button
                        round
                        v-if="appDetail.enable && operate === 'install'"
                        @click="openInstall"
                        type="primary"
                        class="brief-button"
                    >
                        {{ $t('app.install') }}
                    </el-button>
                </div>
            </div>
            <div class="divider"></div>
            <div class="descriptions">
                <div>
                    <el-descriptions direction="vertical">
                        <el-descriptions-item>
                            <div class="icons">
                                <el-link @click="toLink(app.website)">
                                    <el-icon><OfficeBuilding /></el-icon>
                                    <span>{{ $t('app.appOfficeWebsite') }}</span>
                                </el-link>
                            </div>
                        </el-descriptions-item>
                        <el-descriptions-item>
                            <el-link @click="toLink(app.document)">
                                <el-icon><Document /></el-icon>
                                <span>{{ $t('app.document') }}</span>
                            </el-link>
                        </el-descriptions-item>
                        <el-descriptions-item>
                            <el-link @click="toLink(app.github)">
                                <el-icon><Link /></el-icon>
                                <span>{{ $t('app.github') }}</span>
                            </el-link>
                        </el-descriptions-item>
                    </el-descriptions>
                </div>
            </div>
        </div>
        <MdEditor previewOnly v-model="app.readMe" :theme="isDarkTheme ? 'dark' : 'light'" />
    </el-drawer>
    <Install ref="installRef"></Install>
</template>

<script lang="ts" setup>
import { GetApp, GetAppDetail } from '@/api/modules/app';
import MdEditor from 'md-editor-v3';
import { ref } from 'vue';
import Install from './install/index.vue';
import router from '@/routers';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);

const app = ref<any>({});
const appDetail = ref<any>({});
const version = ref('');
const loadingDetail = ref(false);
const loadingApp = ref(false);
const installRef = ref();
const open = ref(false);
const appKey = ref();
const operate = ref();

const acceptParams = async (key: string, op: string) => {
    appKey.value = key;
    operate.value = op;
    open.value = true;
    getApp();
};

const handleClose = () => {
    open.value = false;
};

const getApp = async () => {
    loadingApp.value = true;
    try {
        const res = await GetApp(appKey.value);
        app.value = res.data;
        app.value.icon = 'data:image/png;base64,' + res.data.icon;
        version.value = app.value.versions[0];
        getDetail(app.value.id, version.value);
    } finally {
        loadingApp.value = false;
    }
};

const getDetail = async (id: number, version: string) => {
    loadingDetail.value = true;
    try {
        const res = await GetAppDetail(id, version, 'app');
        appDetail.value = res.data;
    } finally {
        loadingDetail.value = false;
    }
};

const toLink = (link: string) => {
    window.open(link, '_blank');
};

const openInstall = () => {
    switch (app.value.type) {
        case 'php':
        case 'node':
        case 'java':
        case 'go':
        case 'python':
        case 'dotnet':
            router.push({ path: '/websites/runtimes/' + app.value.type });
            break;
        default:
            const params = {
                app: app.value,
            };
            installRef.value.acceptParams(params);
            open.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.brief {
    .name {
        span {
            font-weight: 500;
            font-size: 18px;
            color: var(--el-text-color-regular);
        }
    }

    .description {
        margin-top: 10px;
        span {
            font-size: 14px;
            color: var(--el-text-color-regular);
        }
    }

    .icon {
        width: 180px;
        height: 180px;
        background-color: #ffffff;
    }

    .version {
        margin-top: 10px;
    }

    .descriptions {
        margin-top: 5px;
        .icons {
            margin-left: 20px;
        }
    }
}

:deep(.md-editor-dark) {
    background-color: var(--panel-main-bg-color-9);
}
</style>
