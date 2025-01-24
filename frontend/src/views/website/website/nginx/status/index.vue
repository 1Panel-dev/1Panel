<template>
    <div>
        <el-form label-position="top">
            <el-row type="flex" class="ml-5" justify="center">
                <el-form-item class="w-1/4">
                    <template #label>
                        <span class="status-label">{{ $t('nginx.connections') }}</span>
                    </template>
                    <span class="status-count">{{ data.active }}</span>
                </el-form-item>
                <el-form-item class="w-1/4">
                    <template #label>
                        <span class="status-label">{{ $t('nginx.accepts') }}</span>
                    </template>
                    <span class="status-count">{{ data.accepts }}</span>
                </el-form-item>
                <el-form-item class="w-1/4">
                    <template #label>
                        <span class="status-label">{{ $t('nginx.handled') }}</span>
                    </template>
                    <span class="status-count">{{ data.handled }}</span>
                </el-form-item>
                <el-form-item class="w-1/4">
                    <template #label>
                        <span class="status-label">{{ $t('nginx.requests') }}</span>
                    </template>
                    <span class="status-count">{{ data.requests }}</span>
                </el-form-item>
                <el-form-item class="w-1/4">
                    <template #label>
                        <span class="status-label">{{ $t('nginx.reading') }}</span>
                    </template>
                    <span class="status-count">{{ data.reading }}</span>
                </el-form-item>
                <el-form-item class="w-1/4">
                    <template #label>
                        <span class="status-label">{{ $t('nginx.writing') }}</span>
                    </template>
                    <span class="status-count">{{ data.writing }}</span>
                </el-form-item>
                <el-form-item class="w-1/4">
                    <template #label>
                        <span class="status-label">{{ $t('nginx.waiting') }}</span>
                    </template>
                    <span class="status-count">{{ data.waiting }}</span>
                </el-form-item>
                <el-form-item class="w-1/4" />
            </el-row>
        </el-form>
    </div>
</template>

<script lang="ts" setup>
import { Nginx } from '@/api/interface/nginx';
import { getNginxStatus } from '@/api/modules/nginx';
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
    const res = await getNginxStatus();
    data.value = res.data;
};

onMounted(() => {
    get();
});
</script>
