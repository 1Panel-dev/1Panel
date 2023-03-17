<template>
    <el-row :gutter="20">
        <el-col :span="14" :offset="1">
            <br />
            <el-descriptions :column="1" border v-loading="loading">
                <el-descriptions-item :label="$t('website.siteAlias')">{{ website.alias }}</el-descriptions-item>
                <el-descriptions-item :label="$t('website.primaryPath')">
                    {{ website.sitePath }}
                    <el-button type="primary" link @click="toFolder(website.sitePath)">
                        <el-icon><FolderOpened /></el-icon>
                    </el-button>
                </el-descriptions-item>
            </el-descriptions>
            <br />

            <el-descriptions :title="$t('website.folderTitle')" :column="1" border>
                <el-descriptions-item label="waf">{{ $t('website.wafFolder') }}</el-descriptions-item>
                <el-descriptions-item label="ssl">{{ $t('website.sslFolder') }}</el-descriptions-item>
                <el-descriptions-item label="log">{{ $t('website.logFoler') }}</el-descriptions-item>
                <el-descriptions-item label="index">{{ $t('website.indexFolder') }}</el-descriptions-item>
            </el-descriptions>
        </el-col>
    </el-row>
</template>
<script lang="ts" setup>
import { GetWebsite } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
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
let website = ref<any>({});
let loading = ref(false);

const search = () => {
    loading.value = true;
    GetWebsite(websiteId.value)
        .then((res) => {
            website.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
};

const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};

onMounted(() => {
    search();
});
</script>
