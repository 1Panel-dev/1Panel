<template>
    <div>
        <el-row>
            <el-col :xs="24" :sm="6" :md="6" :lg="6" :xl="6">
                <el-statistic :title="$t('nginx.connections')" :value="data.active" />
            </el-col>
            <el-col :xs="24" :sm="6" :md="6" :lg="6" :xl="6">
                <el-statistic :title="$t('nginx.accepts')" :value="data.accepts" />
            </el-col>
            <el-col :xs="24" :sm="6" :md="6" :lg="6" :xl="6">
                <el-statistic :title="$t('nginx.handled')" :value="data.handled" />
            </el-col>
            <el-col :xs="24" :sm="6" :md="6" :lg="6" :xl="6">
                <el-statistic :title="$t('nginx.requests')" :value="data.requests" />
            </el-col>
            <el-col :xs="24" :sm="6" :md="6" :lg="6" :xl="6">
                <el-statistic :title="$t('nginx.reading')" :value="data.reading" />
            </el-col>
            <el-col :xs="24" :sm="6" :md="6" :lg="6" :xl="6">
                <el-statistic :title="$t('nginx.writing')" :value="data.writing" />
            </el-col>
            <el-col :xs="24" :sm="6" :md="6" :lg="6" :xl="6">
                <el-statistic :title="$t('nginx.waiting')" :value="data.waiting" />
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import { Nginx } from '@/api/interface/nginx';
import { GetNginxStatus } from '@/api/modules/nginx';
import { onMounted, ref } from 'vue';

const props = defineProps({
    status: {
        type: String,
        default: 'Running',
    },
});

let data = ref<Nginx.NginxStatus>({
    accepts: 0,
    handled: 0,
    requests: 0,
    reading: 0,
    waiting: 0,
    writing: 0,
    active: 0,
});

const get = async () => {
    if (props.status != 'Running') {
        return;
    }
    const res = await GetNginxStatus();
    data.value = res.data;
};

onMounted(() => {
    get();
});
</script>

<style scoped>
.el-col {
    text-align: center;
}
</style>
