<template>
    <div>
        <el-card v-if="isActive !== null && !isActive" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { loadDockerStatus } from '@/api/modules/container';
import router from '@/routers';

const em = defineEmits(['search', 'mounted', 'update:is-active', 'update:loading']);
const isActive = ref(null);
const loadStatus = async () => {
    em('update:loading', true);
    await loadDockerStatus()
        .then((res) => {
            isActive.value = res.data === 'Running';
            em('update:loading', false);
            em('update:is-active', isActive.value);
            em('search');
            em('mounted');
        })
        .catch(() => {
            em('update:loading', false);
            em('update:is-active', false);
        });
};

const goSetting = async () => {
    router.push({ name: 'ContainerSetting' });
};

onMounted(() => {
    loadStatus();
});
</script>
