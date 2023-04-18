<template>
    <div v-loading="loading">
        <el-row :gutter="20">
            <el-col :span="14" :offset="1">
                <br />
                <el-descriptions :column="1" border>
                    <el-descriptions-item :label="$t('website.siteAlias')">{{ website.alias }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('website.primaryPath')">
                        {{ website.sitePath }}
                        <el-button type="primary" link @click="toFolder(website.sitePath)">
                            <el-icon><FolderOpened /></el-icon>
                        </el-button>
                    </el-descriptions-item>
                </el-descriptions>
                <br />
                <el-alert :closable="false">
                    <template #default>
                        <span class="warnHelper">{{ $t('website.runUserHelper') }}</span>
                    </template>
                </el-alert>
                <br />
                <el-descriptions :title="$t('website.folderTitle')" :column="1" border>
                    <el-descriptions-item label="waf">{{ $t('website.wafFolder') }}</el-descriptions-item>
                    <el-descriptions-item label="ssl">{{ $t('website.sslFolder') }}</el-descriptions-item>
                    <el-descriptions-item label="log">{{ $t('website.logFoler') }}</el-descriptions-item>
                    <el-descriptions-item label="index">{{ $t('website.indexFolder') }}</el-descriptions-item>
                </el-descriptions>
            </el-col>
            <el-col :span="14" :offset="1" v-if="configDir">
                <br />
                <el-form :inline="true" ref="siteForm" :model="update">
                    <el-form-item :label="$t('website.runDir')" prop="runDir">
                        <el-select v-model="update.siteDir">
                            <el-option :label="'/'" :value="'/'"></el-option>
                            <el-option
                                v-for="(item, index) in dirs"
                                :label="item"
                                :value="item"
                                :key="index"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="submit(siteForm)">{{ $t('nginx.saveAndReload') }}</el-button>
                    </el-form-item>
                </el-form>
                <el-form-item>
                    <el-alert :closable="false">
                        <template #default>
                            <span class="warnHelper">{{ $t('website.runDirHelper') }}</span>
                        </template>
                    </el-alert>
                </el-form-item>
            </el-col>
        </el-row>
    </div>
</template>
<script lang="ts" setup>
import { GetFilesList } from '@/api/modules/files';
import { GetWebsite, UpdateWebsiteDir } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const websiteId = computed(() => {
    return Number(props.id);
});
const website = ref<any>({});
const loading = ref(false);
const configDir = ref(false);
const update = reactive({
    id: 0,
    siteDir: '/',
});
const siteForm = ref<FormInstance>();
const dirReq = reactive({
    path: '/',
    expand: true,
    showHidden: false,
    page: 1,
    pageSize: 100,
    search: '',
    containSub: false,
    dir: true,
});
const dirs = ref([]);

const search = () => {
    loading.value = true;
    GetWebsite(websiteId.value)
        .then((res) => {
            website.value = res.data;
            update.id = website.value.id;
            update.siteDir = website.value.siteDir;
            if (website.value.type === 'static' || website.value.runtimeID > 0) {
                configDir.value = true;
                dirReq.path = website.value.sitePath + '/index';
                getDirs();
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        UpdateWebsiteDir(update)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                search();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const getDirs = async () => {
    loading.value = true;
    await GetFilesList(dirReq)
        .then((res) => {
            const items = res.data.items || [];
            for (const item of items) {
                dirs.value.push(item.name);
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const initData = () => {
    dirs.value = [];
};

const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};

onMounted(() => {
    initData();
    search();
});
</script>

<style scoped>
.warnHelper {
    white-space: pre-line;
    display: block;
}
</style>
