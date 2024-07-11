<template>
    <div v-loading="loading">
        <div class="site-form-wrapper">
            <el-form class="site-form moblie-form" ref="siteForm" :model="update" label-width="100px">
                <el-form-item :label="$t('website.siteAlias')">
                    {{ website.alias }}
                </el-form-item>
                <el-form-item :label="$t('website.primaryPath')">
                    <el-space wrap>
                        {{ website.sitePath + '/index' }}
                        <el-button type="primary" link @click="toFolder(website.sitePath + '/index')">
                            <el-icon>
                                <FolderOpened />
                            </el-icon>
                        </el-button>
                    </el-space>
                    <span class="input-help" v-if="configDir">
                        {{ $t('php.indexHelper') }}
                    </span>
                </el-form-item>
                <el-form-item v-if="configDir" :label="$t('website.runDir')">
                    <el-space wrap>
                        <el-select v-model="update.siteDir" filterable class="p-w-200">
                            <el-option
                                v-for="(item, index) in dirs"
                                :label="item"
                                :value="item"
                                :key="index"
                            ></el-option>
                        </el-select>
                        <el-button type="primary" @click="submit(siteForm)">
                            {{ $t('nginx.saveAndReload') }}
                        </el-button>
                    </el-space>
                    <span class="input-help">
                        {{ $t('website.runDirHelper2') }}
                    </span>
                </el-form-item>
                <el-form-item v-if="configDir" :label="$t('website.userGroup')">
                    <el-space wrap>
                        <el-input v-model="updatePermission.user" class="user-num-input">
                            <template #prepend>{{ $t('commons.table.user') }}</template>
                        </el-input>
                        <el-input v-model="updatePermission.group" class="user-num-input">
                            <template #prepend>{{ $t('website.uGroup') }}</template>
                        </el-input>
                        <el-button type="primary" @click="submitPermission()">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </el-space>
                </el-form-item>
            </el-form>
            <el-text type="warning" v-if="configDir">{{ $t('website.runUserHelper') }}</el-text>
            <br />
            <el-text type="danger" v-if="dirConfig.msg != ''">{{ dirConfig.msg }}</el-text>
            <br />
            <el-descriptions :title="$t('website.folderTitle')" :column="1" border>
                <el-descriptions-item label="ssl">{{ $t('website.sslFolder') }}</el-descriptions-item>
                <el-descriptions-item label="log">{{ $t('website.logFolder') }}</el-descriptions-item>
                <el-descriptions-item label="index">{{ $t('website.indexFolder') }}</el-descriptions-item>
            </el-descriptions>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { GetDirConfig, GetWebsite, UpdateWebsiteDir, UpdateWebsiteDirPermission } from '@/api/modules/website';
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
const updatePermission = reactive({
    id: 0,
    user: '1000',
    group: '1000',
});
const siteForm = ref<FormInstance>();
const dirs = ref([]);
const dirConfig = ref<Website.DirConfig>({
    dirs: [''],
    user: '',
    userGroup: '',
    msg: '',
});

const search = () => {
    loading.value = true;
    GetWebsite(websiteId.value)
        .then((res) => {
            website.value = res.data;
            update.id = website.value.id;
            update.siteDir = website.value.siteDir.startsWith('/')
                ? website.value.siteDir
                : '/' + website.value.siteDir;
            updatePermission.id = website.value.id;
            updatePermission.group = website.value.group === '' ? '1000' : website.value.group;
            updatePermission.user = website.value.user === '' ? '1000' : website.value.user;
            if (website.value.type === 'static' || website.value.runtimeID > 0) {
                configDir.value = true;
                getDirConfig();
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

const submitPermission = async () => {
    if (updatePermission.user === '' || updatePermission.group === '') {
        return;
    }
    loading.value = true;
    UpdateWebsiteDirPermission(updatePermission)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const initData = () => {
    dirs.value = [];
};

function filterDirectories(directories: any[]) {
    return directories.filter((dir) => {
        return dir !== '/node_modules' && dir !== '/vendor';
    });
}

const getDirConfig = async () => {
    try {
        const res = await GetDirConfig({ id: props.id });
        dirs.value = res.data.dirs;
        dirs.value = filterDirectories(dirs.value);
        dirConfig.value = res.data;
    } catch (error) {}
};

const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};

onMounted(() => {
    initData();
    search();
});
</script>

<style lang="scss" scoped>
.site-form-wrapper {
    min-width: 600px;
    width: 60%;
    padding: 20px;
}
.site-form {
    :deep(.el-form-item__label) {
        padding-right: 20px !important;
        box-sizing: content-box;
    }
    .user-num-input {
        width: 190px;
    }
}
.warnHelper {
    white-space: pre-line;
    display: block;
}
</style>
