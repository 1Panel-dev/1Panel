<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.websiteLog')">
            <template #toolbar>
                <div class="flex w-full flex-col gap-4 md:justify-between md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-button
                            class="tag-button"
                            :class="logConfig.name === 'access.log' ? '' : 'no-active'"
                            :type="logConfig.name === 'access.log' ? 'primary' : ''"
                            @click="changeType('access.log')"
                        >
                            {{ $t('logs.runLog') }}
                        </el-button>
                        <el-button
                            class="tag-button"
                            :class="logConfig.name === 'error.log' ? '' : 'no-active'"
                            :type="logConfig.name === 'error.log' ? 'primary' : ''"
                            @click="changeType('error.log')"
                        >
                            {{ $t('logs.errLog') }}
                        </el-button>
                    </div>
                </div>
            </template>
            <template #search>
                <div class="flex flex-wrap items-center gap-2 sm:gap-4">
                    <el-select v-model="logConfig.id" @change="changeWebsite()" class="p-w-200">
                        <template #prefix>{{ $t('website.website') }}</template>
                        <el-option
                            v-for="(website, index) in websites"
                            :key="index"
                            :label="website.primaryDomain"
                            :value="website.id"
                        ></el-option>
                    </el-select>
                    <el-button>
                        <el-checkbox v-model="tailLog" @change="changeTail" :disabled="logConfig.id == undefined">
                            {{ $t('commons.button.watch') }}
                        </el-checkbox>
                    </el-button>
                    <el-button @click="onDownload" icon="Download" :disabled="!hasContent">
                        {{ $t('file.download') }}
                    </el-button>
                    <el-button type="primary" plain @click="onClean()" :disabled="!hasContent">
                        {{ $t('logs.deleteLogs') }}
                    </el-button>
                </div>
            </template>
            <template #main>
                <LogFile
                    ref="logRef"
                    :config="logConfig"
                    :default-button="false"
                    v-if="showLog"
                    v-model:loading="loading"
                    v-model:hasContent="hasContent"
                    :style="'height: calc(100vh - 370px)'"
                />
            </template>
        </LayoutContent>
        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitClean"></ConfirmDialog>
    </div>
</template>
<script setup lang="ts">
import { ListWebsites, OpWebsiteLog } from '@/api/modules/website';
import { reactive } from 'vue';
import { onMounted } from 'vue';
import { ref, nextTick } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import LogFile from '@/components/log-file/index.vue';

const logConfig = reactive({
    type: 'website',
    id: undefined,
    name: 'access.log',
});
const showLog = ref(false);
const loading = ref(false);
const websites = ref();
const confirmDialogRef = ref();
const tailLog = ref(false);
const logRef = ref();
const hasContent = ref(false);

const searchLog = () => {
    showLog.value = false;
    nextTick(() => {
        showLog.value = true;
    });
};

const getWebsites = async () => {
    loading.value = true;
    await ListWebsites()
        .then((res) => {
            websites.value = res.data || [];
            if (websites.value.length > 0) {
                logConfig.id = websites.value[0].id;
                showLog.value = true;
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const changeType = (type: string) => {
    logConfig.name = type;
    if (logConfig.id != undefined) {
        searchLog();
    }
};

const changeWebsite = () => {
    searchLog();
};

const onClean = async () => {
    const params = {
        header: i18n.global.t('logs.deleteLogs'),
        operationInfo: i18n.global.t('commons.msg.delete'),
        submitInputInfo: i18n.global.t('logs.deleteLogs'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onDownload = async () => {
    logRef.value.onDownload();
};

const changeTail = () => {
    logRef.value.changeTail(true);
};

// const onCloseLog = async () => {
//     tailLog.value = false;
//     clearInterval(Number(timer));
//     timer = null;
// };

const onSubmitClean = async () => {
    const req = {
        id: logConfig.id,
        operate: 'delete',
        logType: logConfig.name,
    };
    loading.value = true;
    OpWebsiteLog(req)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            searchLog();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    getWebsites();
});
</script>

<style scoped>
.el-button + .el-button {
    margin: 0 !important;
}
</style>
