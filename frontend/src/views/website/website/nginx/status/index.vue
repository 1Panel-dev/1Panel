<template>
    <el-row>
        <el-col :span="22" :offset="1">
            <el-descriptions :column="4" direction="vertical">
                <el-descriptions-item>
                    <template #label>
                        <span class="status-label">{{ $t('nginx.connections') }}</span>
                    </template>
                    <span class="status-count">{{ data.active }}</span>
                </el-descriptions-item>
                <el-descriptions-item>
                    <template #label>
                        <span class="status-label">{{ $t('nginx.accepts') }}</span>
                    </template>
                    <span class="status-count">{{ data.accepts }}</span>
                </el-descriptions-item>
                <el-descriptions-item>
                    <template #label>
                        <span class="status-label">{{ $t('nginx.handled') }}</span>
                    </template>
                    <span class="status-count">{{ data.handled }}</span>
                </el-descriptions-item>
                <el-descriptions-item>
                    <template #label>
                        <span class="status-label">{{ $t('nginx.requests') }}</span>
                    </template>
                    <span class="status-count">{{ data.requests }}</span>
                </el-descriptions-item>
                <el-descriptions-item>
                    <template #label>
                        <span class="status-label">{{ $t('nginx.reading') }}</span>
                    </template>
                    <span class="status-count">{{ data.reading }}</span>
                </el-descriptions-item>
                <el-descriptions-item>
                    <template #label>
                        <span class="status-label">{{ $t('nginx.writing') }}</span>
                    </template>
                    <span class="status-count">{{ data.writing }}</span>
                </el-descriptions-item>
                <el-descriptions-item>
                    <template #label>
                        <span class="status-label">{{ $t('nginx.waiting') }}</span>
                    </template>
                    <span class="status-count">{{ data.waiting }}</span>
                </el-descriptions-item>
            </el-descriptions>
        </el-col>
    </el-row>
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
    accepts: '',
    handled: '',
    requests: '',
    reading: '',
    waiting: '',
    writing: '',
    active: '',
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

<style lang="scss" scoped>
.status-count {
    font-size: 24px;
}

.status-label {
    font-size: 14px;
    color: #646a73;
}
</style>
