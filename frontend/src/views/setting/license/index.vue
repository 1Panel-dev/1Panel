<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.license')" :divider="true">
            <template #main>
                <el-row :gutter="20" class="mt-5; mb-10">
                    <el-col :xs="24" :sm="24" :md="15" :lg="15" :xl="15">
                        <div class="descriptions">
                            <el-descriptions :column="1" direction="horizontal" size="large" border>
                                <el-descriptions-item :label="$t('license.authorizationId')">
                                    {{ license.licenseName || '-' }}
                                    <el-button type="primary" class="ml-3" plain @click="onSync" size="small">
                                        {{ $t('commons.button.sync') }}
                                    </el-button>
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.authorizedUser')">
                                    {{ license.assigneeName || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.expiresAt')">
                                    {{ license.expiresAt || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.productName')">
                                    {{ license.productName || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.productStatus')">
                                    <div v-if="license.status">
                                        <el-tooltip
                                            v-if="license.status.indexOf('lost') !== -1"
                                            :content="$t('license.lostHelper')"
                                        >
                                            <el-tag type="info">
                                                {{ $t('license.' + license.status) }}
                                            </el-tag>
                                        </el-tooltip>
                                        <el-tag v-else>{{ $t('license.' + license.status) }}</el-tag>
                                    </div>
                                    <span v-else>-</span>
                                </el-descriptions-item>
                            </el-descriptions>
                        </div>
                    </el-col>

                    <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                        <CardWithHeader :header="$t('license.quickUpdate')">
                            <template #body>
                                <div class="h-app-card">
                                    <el-row :gutter="10">
                                        <el-col :span="15">
                                            <div class="h-app-content">{{ $t('license.importLicense') }}：</div>
                                        </el-col>
                                        <el-col :span="5">
                                            <el-button type="primary" plain round size="small" @click="toUpload">
                                                {{ $t('license.import') }}
                                            </el-button>
                                        </el-col>
                                    </el-row>
                                </div>
                                <div class="h-app-card">
                                    <el-row :gutter="10">
                                        <el-col :span="15">
                                            <div class="h-app-content">{{ $t('license.technicalAdvice') }}：</div>
                                        </el-col>
                                        <el-col :span="5">
                                            <el-button type="primary" plain round size="small" @click="toHalo()">
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

        <Upload ref="uploadRef" @search="search()" />
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { getLicense, syncLicense } from '@/api/modules/setting';
import CardWithHeader from '@/components/card-with-header/index.vue';
import Upload from '@/views/setting/license/upload/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
const loading = ref();
const uploadRef = ref();

const license = reactive({
    licenseName: '',
    trial: true,
    expiresAt: '',
    assigneeName: '',
    productName: '',

    status: '',
});

const toHalo = () => {
    window.open('https://halo.test.lxware.cn/', '_blank', 'noopener,noreferrer');
};

const onSync = async () => {
    loading.value = true;
    await syncLicense()
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
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

const search = async () => {
    loading.value = true;
    await getLicense()
        .then((res) => {
            loading.value = false;
            license.status = res.data.status;
            if (res.data.status !== 'Enable') {
                return;
            }
            license.licenseName = res.data.licenseName;
            license.assigneeName = res.data.assigneeName;
            license.trial = res.data.trial;
            if (res.data.productPro) {
                license.productName = 'product-1panel-pro';
                license.expiresAt =
                    res.data.productPro === '0'
                        ? i18n.global.t('license.indefinitePeriod')
                        : timestampToDate(Number(res.data.productPro));
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const toUpload = () => {
    uploadRef.value.acceptParams();
};

onMounted(() => {
    search();
});
</script>

<style scoped lang="scss">
.h-app-card {
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
</style>
