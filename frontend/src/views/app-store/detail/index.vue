<template>
    <DrawerPro v-model="open" :header="$t('app.detail')" :back="handleClose" size="large">
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
                            {{ language == 'zh' || language == 'tw' ? app.shortDescZh : app.shortDescEn }}
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
                        {{ $t('commons.button.install') }}
                    </el-button>
                </div>
            </div>
            <div class="descriptions">
                <el-descriptions border size="large" direction="vertical">
                    <el-descriptions-item :label="$t('app.appOfficeWebsite')">
                        <el-link @click="toLink(app.website)">
                            {{ $t('app.link') }}
                            <el-icon class="ml-1.5"><Promotion /></el-icon>
                        </el-link>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('app.github')">
                        <el-link @click="toLink(app.github)">
                            {{ $t('app.link') }}
                            <el-icon class="ml-1.5"><Promotion /></el-icon>
                        </el-link>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('app.requireMemory')" v-if="appDetail.memoryRequired > 0">
                        <span>{{ computeSizeFromMB(appDetail.memoryRequired) }}</span>
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('app.supportedArchitectures')" v-if="architectures.length > 0">
                        <el-tag v-for="(arch, index) in architectures" :key="index" class="mx-1">
                            {{ arch }}
                        </el-tag>
                    </el-descriptions-item>
                </el-descriptions>
            </div>
        </div>
        <MdEditor previewOnly v-model="app.readMe" :theme="isDarkTheme ? 'dark' : 'light'" />
    </DrawerPro>
    <Install ref="installRef"></Install>
</template>

<script lang="ts" setup>
import { getAppByKey, getAppDetail } from '@/api/modules/app';
import MdEditor from 'md-editor-v3';
import { ref } from 'vue';
import Install from './install/index.vue';
import router from '@/routers';
import { GlobalStore } from '@/store';
import { getLanguage, computeSizeFromMB } from '@/utils/util';
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);

const language = getLanguage();

const app = ref<any>({});
const appDetail = ref<any>({});
const version = ref('');
const loadingDetail = ref(false);
const loadingApp = ref(false);
const installRef = ref();
const open = ref(false);
const appKey = ref();
const operate = ref();
const architectures = ref([]);

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
        const res = await getAppByKey(appKey.value);
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
        const res = await getAppDetail(id, version, 'app');
        appDetail.value = res.data;
        if (appDetail.value.architectures != '') {
            architectures.value = appDetail.value.architectures.split(',');
        }
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
            router.push({ path: '/websites/runtimes/php' });
            break;
        case 'node':
            router.push({ path: '/websites/runtimes/node' });
            break;
        case 'java':
            router.push({ path: '/websites/runtimes/java' });
            break;
        case 'go':
            router.push({ path: '/websites/runtimes/go' });
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

<style scoped lang="scss">
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
        margin-top: 20px;
    }
}
:deep(.md-editor-dark) {
    background-color: var(--panel-main-bg-color-9);
}
</style>
