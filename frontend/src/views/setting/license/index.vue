<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('setting.license')" :divider="true">
            <template #main>
                <el-row :gutter="20" class="mt-5; mb-10">
                    <el-col :xs="24" :sm="24" :md="15" :lg="15" :xl="15">
                        <div class="descriptions" v-if="hasLicense">
                            <el-descriptions :column="1" direction="horizontal" size="large" border>
                                <el-descriptions-item :label="$t('license.authorizationId')">
                                    {{ license.licenseName || '-' }}
                                    <el-button
                                        type="primary"
                                        class="ml-3"
                                        plain
                                        @click="onSync"
                                        size="small"
                                        v-if="showSync()"
                                    >
                                        {{ $t('commons.button.sync') }}
                                    </el-button>
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.authorizedUser')">
                                    {{ license.assigneeName || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.productName')">
                                    {{ license.productName || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.trialInfo')">
                                    {{ license.trial ? $t('license.trial') : $t('license.office') }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.expiresAt')">
                                    {{ license.expiresAt || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('license.productStatus')">
                                    <div v-if="license.status">
                                        <el-tooltip
                                            v-if="license.status.indexOf('Lost') !== -1"
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

                        <CardWithHeader :header="$t('home.overview')" height="160px" v-if="!hasLicense">
                            <template #body>
                                <div class="h-app-card">
                                    <el-row>
                                        <el-col :span="6">
                                            <span>{{ $t('setting.license') }}</span>
                                        </el-col>
                                        <el-col :span="6">
                                            <span>{{ $t('license.community2') }}</span>
                                        </el-col>
                                    </el-row>
                                </div>
                            </template>
                        </CardWithHeader>
                    </el-col>

                    <el-col :xs="24" :sm="24" :md="9" :lg="9" :xl="9">
                        <CardWithHeader :header="$t('license.quickUpdate')" height="160px">
                            <template #body>
                                <div class="h-app-card">
                                    <el-row>
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
                                    <el-row>
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

        <LicenseImport ref="licenseRef" />
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { getLicense, syncLicense } from '@/api/modules/setting';
import CardWithHeader from '@/components/card-with-header/index.vue';
import LicenseImport from '@/components/license-import/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const loading = ref();
const licenseRef = ref();
const globalStore = GlobalStore();
const hasLicense = ref();

const license = reactive({
    licenseName: '',
    trial: true,
    expiresAt: '',
    assigneeName: '',
    productName: '',

    status: '',
});

const toHalo = () => {
    window.open('https://www.lxware.cn/1panel' + '', '_blank', 'noopener,noreferrer');
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
    const date = new Date(timestamp * 1000);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${y}-${m}-${d} ${h}:${minute}:${second}`;
};

const search = async () => {
    loading.value = true;
    await getLicense()
        .then((res) => {
            loading.value = false;
            license.status = res.data.status;
            globalStore.isProductPro =
                res.data.status === 'Enable' || res.data.status === 'Lost01' || res.data.status === 'Lost02';
            if (res.data.status === '') {
                hasLicense.value = false;
                return;
            }
            hasLicense.value = true;
            if (globalStore.isProductPro) {
                globalStore.productProExpires = Number(res.data.productPro);
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

const showSync = () => {
    return license.status.indexOf('Lost') !== -1 || license.status === 'Disable';
};

const toUpload = () => {
    licenseRef.value.acceptParams();
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
    &:hover {
        background-color: rgba(0, 94, 235, 0.03);
    }
}
</style>
