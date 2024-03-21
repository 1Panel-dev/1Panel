<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.license')" :divider="true">
            <template #main>
                <el-alert style="margin-top: 20px" type="warning" @close="hideEntrance" v-if="show == true">
                    <template #title>
                        <span class="flx-align-center">
                            <span>{{ $t('license.importLicense') }}</span>
                        </span>
                    </template>
                </el-alert>
                <el-row :gutter="20" style="margin-top: 20px">
                    <el-col :xs="24" :sm="24" :md="16" :lg="16" :xl="16">
                        <CardWithHeader :header="$t('home.overview')" height="146px" v-if="show == true">
                            <template #body>
                                <div class="h-overview">
                                    <el-row>
                                        <el-col :span="6">
                                            <span>{{ $t('setting.license') }}</span>
                                        </el-col>
                                        <el-col :span="6">
                                            <span>{{ $t('license.community') }}</span>
                                        </el-col>
                                    </el-row>
                                </div>
                            </template>
                        </CardWithHeader>

                        <CardWithHeader :header="$t('home.systemInfo')" v-if="show == false">
                            <template #body>
                                <el-scrollbar>
                                    <div class="h-overview">
                                        <el-row>
                                            <el-col :span="6">
                                                <span>{{ $t('license.authorizationId') }}</span>
                                            </el-col>
                                            <el-col :span="6">
                                                <span>{{ license.licenseName }}</span>
                                            </el-col>
                                        </el-row>
                                        <div class="h-app-divider" />
                                        <el-row>
                                            <el-col :span="6">
                                                <span>{{ $t('license.authorizedUser') }}</span>
                                            </el-col>
                                            <el-col :span="6">
                                                <span>{{ license.assigneeName }}</span>
                                            </el-col>
                                        </el-row>
                                        <div class="h-app-divider" />
                                        <el-row>
                                            <el-col :span="6">
                                                <span>{{ $t('license.expiresAt') }}</span>
                                            </el-col>
                                            <el-col :span="6">
                                                <span>{{ license.expiresAt }}</span>
                                            </el-col>
                                        </el-row>
                                        <div class="h-app-divider" />
                                        <el-row>
                                            <el-col :span="6">
                                                <span>{{ $t('license.versionName') }}</span>
                                            </el-col>
                                            <el-col :span="6">
                                                <span>{{ license.productName }}</span>
                                            </el-col>
                                        </el-row>
                                        <div class="h-app-divider" />
                                        <el-row>
                                            <el-col :span="6">
                                                <span>{{ $t('license.versionType') }}</span>
                                            </el-col>
                                            <el-col :span="6">
                                                <span>{{ $t('license.pro') }}</span>
                                            </el-col>
                                        </el-row>
                                    </div>
                                </el-scrollbar>
                            </template>
                        </CardWithHeader>
                    </el-col>

                    <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
                        <CardWithHeader :header="$t('license.quickUpdate')">
                            <template #body>
                                <div class="h-app-card">
                                    <el-row :gutter="10">
                                        <el-col :span="15">
                                            <div class="h-app-content">
                                                <div>
                                                    <span class="h-app-title">{{ $t('license.importLicense') }}：</span>
                                                </div>
                                            </div>
                                        </el-col>
                                        <el-col :span="5">
                                            <el-upload
                                                action="#"
                                                ref="uploadRef"
                                                class="upload-demo"
                                                :limit="1"
                                                :on-exceed="handleExceed"
                                                :on-change="fileOnChange"
                                                :auto-upload="false"
                                                v-model:file-list="uploaderFiles"
                                            >
                                                <template #trigger>
                                                    <el-button
                                                        v-if="license.licenseName === ''"
                                                        class="h-app-button"
                                                        type="primary"
                                                        plain
                                                        round
                                                        size="small"
                                                    >
                                                        {{ $t('license.import') }}
                                                    </el-button>
                                                    <el-button
                                                        v-else
                                                        class="h-app-button"
                                                        type="primary"
                                                        plain
                                                        round
                                                        size="small"
                                                    >
                                                        {{ $t('license.update') }}
                                                    </el-button>
                                                </template>
                                            </el-upload>
                                        </el-col>
                                    </el-row>
                                </div>
                                <div class="h-app-card">
                                    <el-row :gutter="10">
                                        <el-col :span="15">
                                            <div class="h-app-content">
                                                <div>
                                                    <span class="h-app-title">
                                                        {{ $t('license.technicalAdvice') }}：
                                                    </span>
                                                </div>
                                            </div>
                                        </el-col>
                                        <el-col :span="5">
                                            <el-button
                                                class="h-app-button"
                                                type="primary"
                                                plain
                                                round
                                                size="small"
                                                @click="toHalo()"
                                            >
                                                {{ $t('license.advice') }}
                                            </el-button>
                                        </el-col>
                                    </el-row>
                                </div>
                            </template>
                        </CardWithHeader>
                    </el-col>
                </el-row>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeMount } from 'vue';
import { GlobalStore } from '@/store';
import CardWithHeader from '@/components/card-with-header/index.vue';
import { genFileId, UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getLicense, UploadFileData } from '@/api/modules/setting';
import i18n from '@/lang';
const globalStore = GlobalStore();
const loading = ref();
const uploadRef = ref<UploadInstance>();
const uploaderFiles = ref<UploadFiles>([]);
const show = ref(null);

const license = reactive({
    licenseName: '',
    trial: true,
    expiresAt: '',
    assigneeName: '',
    productName: '',
});

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const hideEntrance = () => {
    globalStore.setShowEntranceWarn(false);
};

const toHalo = () => {
    window.open('https://halo.test.lxware.cn/', '_blank', 'noopener,noreferrer');
};

const timestampToDate = (timestamp: number) => {
    const date = new Date(timestamp);
    const year = date.getFullYear();
    const month = date.getMonth() + 1;
    const day = date.getDate();
    const hours = date.getHours();
    const minutes = date.getMinutes();
    const seconds = date.getSeconds();
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
};

const get = async () => {
    await getLicense()
        .then((res) => {
            loading.value = false;
            if (res.data !== undefined) {
                license.licenseName = res.data.licenseName;
                license.assigneeName = res.data.assigneeName;
                license.productName = res.data.products[0].name;
                if (res.data.products[0].expiresAt === 0) {
                    license.expiresAt = i18n.global.t('license.indefinitePeriod');
                } else {
                    license.expiresAt = timestampToDate(res.data.products[0].expiresAt);
                }
                show.value = false;
            } else {
                show.value = true;
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const fileOnChange = async (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
    const files = uploaderFiles.value.slice();
    if (files.length !== 1) {
        MsgError(i18n.global.t('license.allowOnlyOne'));
    }
    const file = files[0];
    const formData = new FormData();
    formData.append('file', file.raw);

    await UploadFileData(formData)
        .then(async (res) => {
            if (res) {
                license.licenseName = res.data.licenseName;
                // await get();
                loading.value = false;
                uploadRef.value!.clearFiles();
                uploaderFiles.value = [];
                MsgSuccess(i18n.global.t('license.updateSuccess'));
                window.location.reload(); // 刷新页面
            }
        })
        .catch((error) => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
            MsgError(i18n.global.t(error.message));
        });
};

onMounted(() => {
    getLicense();
});

onBeforeMount(() => {
    get();
});
</script>

<style scoped lang="scss">
.h-app-card {
    cursor: pointer;
    padding: 10px 15px;
    margin-right: 10px;
    line-height: 18px;

    .h-app-content {
        padding-left: 15px;
        .h-app-desc {
            span {
                font-weight: 400;
                font-size: 12px;
                color: var(--el-text-color-regular);
            }
        }
    }
    &:hover {
        background-color: rgba(0, 94, 235, 0.03);
    }
}

.h-app-divider {
    margin-bottom: 13px;
    margin-top: 13px;
    border: 0;
    border-top: var(--panel-border);
}
</style>
