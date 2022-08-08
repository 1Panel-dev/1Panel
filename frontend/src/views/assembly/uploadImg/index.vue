<template>
    <div class="content-box">
        <span class="text">图片上传 🍓🍇🍈🍉</span>
        <UploadImg v-model:imageUrl="avatar" :uploadStyle="uploadStyle">
            <template #tip> 上传成功返回图片是模拟固定的 </template>
        </UploadImg>
    </div>
</template>

<script setup lang="ts" name="uploadImg">
import { ref } from 'vue';
import UploadImg from '@/components/UploadImg/index.vue';

const avatar = ref('');
const uploadStyle = { width: '175px', height: '175px' };
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
