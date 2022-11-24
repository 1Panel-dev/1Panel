<template>
    <el-row>
        <el-col :span="12">
            <el-descriptions :column="1" border>
                <el-descriptions-item :width="100" :label="$t('nginx.connections')">
                    {{ data.active }}
                </el-descriptions-item>
                <el-descriptions-item :label="$t('nginx.accepts')">{{ data.accepts }}</el-descriptions-item>
                <el-descriptions-item :label="$t('nginx.handled')">{{ data.handled }}</el-descriptions-item>
                <el-descriptions-item :label="$t('nginx.requests')">{{ data.requests }}</el-descriptions-item>
                <el-descriptions-item :label="$t('nginx.reading')">{{ data.reading }}</el-descriptions-item>
                <el-descriptions-item :label="$t('nginx.writing')">{{ data.writing }}</el-descriptions-item>
                <el-descriptions-item :label="$t('nginx.waiting')">{{ data.waiting }}</el-descriptions-item>
            </el-descriptions>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { Nginx } from '@/api/interface/nginx';
import { GetNginxStatus } from '@/api/modules/nginx';
import { onMounted, ref } from 'vue';

let data = ref<Nginx.NginxStatus>({
    accepts: '',
    handled: '',
    requests: '',
    reading: '',
    waiting: '',
    writing: '',
    active: '',
});

const get = async () => {
    const res = await GetNginxStatus();
    data.value = res.data;
};

onMounted(() => {
    get();
});
</script>
