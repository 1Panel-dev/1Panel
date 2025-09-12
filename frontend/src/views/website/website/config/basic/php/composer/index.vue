<template>
    <el-form label-position="right" label-width="100px">
        <el-form-item :label="$t('website.execParameters')">
            <el-select v-model="req.command" class="p-w-400">
                <el-option label="install" value="install"></el-option>
                <el-option label="update" value="update"></el-option>
                <el-option label="require" value="require"></el-option>
                <el-option label="create-project" value="create-project"></el-option>
                <el-option :label="$t('container.custom')" value="custom"></el-option>
            </el-select>
        </el-form-item>
        <el-form-item :label="$t('website.extCommand')">
            <el-input v-model.trim="req.extCommand" class="p-w-400"></el-input>
        </el-form-item>
        <el-form-item :label="$t('website.mirror')">
            <el-select v-model="req.mirror" class="p-w-400">
                <el-option
                    v-for="mirror in mirrors"
                    :key="mirror.label"
                    :value="mirror.value"
                    :label="mirror.label + ' [' + mirror.value + ']'"
                ></el-option>
            </el-select>
        </el-form-item>
        <el-form-item :label="$t('website.execUser')">
            <el-select v-model="req.user" class="p-w-400">
                <el-option label="www-data" value="www-data"></el-option>
                <el-option label="root" value="root"></el-option>
            </el-select>
        </el-form-item>
        <el-form-item :label="$t('website.execDir')">
            <el-input v-model.trim="req.dir" class="p-w-400">
                <template #prepend>
                    <el-button icon="Folder" @click="dirRef.acceptParams({ dir: true, path: req.dir })" />
                </template>
            </el-input>
        </el-form-item>
        <el-form-item>
            <el-button type="primary" @click="exec">{{ $t('commons.button.handle') }}</el-button>
        </el-form-item>
    </el-form>
    <TaskLog ref="taskLogRef" @close="search" />
    <FileList ref="dirRef" :dir="true" @choose="getPath" />
</template>
<script setup lang="ts">
import { execComposer, getWebsite } from '@/api/modules/website';
import i18n from '@/lang';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';
import FileList from '@/components/file-list/index.vue';

const props = defineProps({
    websiteID: {
        type: Number,
        default: 0,
    },
});

const req = reactive({
    websiteID: 0,
    command: 'install',
    extCommand: '',
    mirror: 'https://mirrors.aliyun.com/composer/',
    dir: '',
    user: 'www-data',
    taskID: '',
});
const loading = ref(false);
const taskLogRef = ref();
const dirRef = ref();

const mirrors = [
    {
        label: i18n.global.t('runtime.aliyun') + '(mirrors.aliyun.com)',
        value: 'https://mirrors.aliyun.com/composer/',
    },
    {
        label: i18n.global.t('website.tencentCloud') + '(mirrors.cloud.tencent.com)',
        value: 'https://mirrors.cloud.tencent.com/composer/',
    },
    {
        label: i18n.global.t('commons.table.default') + '(repo.packagist.org)',
        value: 'https://repo.packagist.org',
    },
    {
        label: i18n.global.t('website.packagist') + '(packagist.phpcomposer.com)',
        value: 'https://packagist.phpcomposer.com',
    },
    {
        label: i18n.global.t('website.huaweicloud') + '(mirrors.huaweicloud.com)',
        value: 'https://mirrors.huaweicloud.com/repository/php',
    },
    {
        label: 'Packagist Mirror' + '(packagist.mirrors.sjtug.sjtu.edu.cn)',
        value: 'https://packagist.mirrors.sjtug.sjtu.edu.cn/',
    },
];

const getPath = (execDir: string) => {
    req.dir = execDir;
};

const search = () => {
    loading.value = true;
    getWebsite(req.websiteID)
        .then((res) => {
            req.dir = res.data.sitePath + '/index';
        })
        .finally(() => {
            loading.value = false;
        });
};

const exec = async () => {
    const taskID = newUUID();
    req.taskID = taskID;
    try {
        await execComposer(req);
        taskLogRef.value.openWithTaskID(taskID);
    } catch (error) {
        return;
    }
};

onMounted(() => {
    req.websiteID = props.websiteID;
    search();
});
</script>
