<template>
    <LayoutContent :title="$t('app.detail')" :back-name="'App'" :divider="true">
        <template #main>
            <div class="brief" v-loading="loadingApp">
                <el-row :gutter="20">
                    <div>
                        <el-col :span="3">
                            <el-avatar shape="square" :size="180" :src="app.icon" />
                        </el-col>
                    </div>
                    <el-col :span="18">
                        <div class="detail">
                            <div class="name">
                                <span>{{ app.name }}</span>
                            </div>
                            <div class="description">
                                <span>
                                    {{ language == 'zh' || language == 'tw' ? app.shortDescZh : app.shortDescEn }}
                                </span>
                            </div>
                            <div class="version">
                                <el-form-item :label="$t('app.version')">
                                    <el-select v-model="version" @change="getDetail(app.id, version)">
                                        <el-option
                                            v-for="(v, index) in app.versions"
                                            :key="index"
                                            :value="v"
                                            :label="v"
                                        >
                                            {{ v }}
                                        </el-option>
                                    </el-select>
                                </el-form-item>
                            </div>

                            <br />
                            <div v-if="!loadingDetail">
                                <el-alert
                                    style="width: 300px"
                                    v-if="!appDetail.enable"
                                    :title="$t('app.limitHelper')"
                                    type="warning"
                                    show-icon
                                    :closable="false"
                                />
                            </div>
                            <div>
                                <el-button round v-if="appDetail.enable" @click="openInstall" type="primary">
                                    {{ $t('app.install') }}
                                </el-button>
                            </div>
                        </div>
                    </el-col>
                </el-row>
                <div class="divider"></div>
                <div>
                    <el-row>
                        <el-col :span="12">
                            <div class="descriptions">
                                <el-descriptions direction="vertical">
                                    <el-descriptions-item>
                                        <el-link @click="toLink(app.website)">
                                            <el-icon><OfficeBuilding /></el-icon>
                                            <span>{{ $t('app.appOfficeWebsite') }}</span>
                                        </el-link>
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
                        </el-col>
                    </el-row>
                </div>
            </div>
            <div style="margin-left: 10px">
                <MdEditor
                    v-model="app.readMe"
                    previewOnly
                    :theme="globalStore.$state.themeConfig.theme === 'dark' ? 'dark' : 'light'"
                />
            </div>
        </template>
    </LayoutContent>
    <Install ref="installRef"></Install>
</template>

<script lang="ts" setup>
import { GetApp, GetAppDetail } from '@/api/modules/app';
import MdEditor from 'md-editor-v3';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import Install from './install/index.vue';
import router from '@/routers';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const language = useI18n().locale.value;

interface OperateProps {
    appKey: string;
}

const props = withDefaults(defineProps<OperateProps>(), {
    // id: 0,
    appKey: '',
});
const app = ref<any>({});
const appDetail = ref<any>({});
const version = ref('');
const loadingDetail = ref(false);
const loadingApp = ref(false);
const installRef = ref();

const getApp = async () => {
    loadingApp.value = true;
    try {
        const res = await GetApp(props.appKey);
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
    let params = {
        params: appDetail.value.params,
        appDetailId: appDetail.value.id,
        app: app.value,
        compose: appDetail.value.dockerCompose,
    };
    if (app.value.type === 'php') {
        router.push({ path: '/websites/runtime/php' });
    } else {
        installRef.value.acceptParams(params);
    }
};

onMounted(() => {
    getApp();
});
</script>

<style lang="scss">
.brief {
    .name {
        span {
            font-weight: 500;
            font-size: 18px;
        }
    }

    .description {
        margin-top: 10px;
        span {
            font-size: 14px;
            color: var(--el-text-color-regular);
        }
    }

    .version {
        margin-top: 10px;
    }

    .descriptions {
        margin-top: 5px;
    }
}
</style>
