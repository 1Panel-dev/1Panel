<template>
    <div class="content-box">
        <span class="text">长按指令 🍇🍇🍇🍓🍓🍓</span>
        <el-button type="primary" v-longpress="longpress"
            >长按2秒触发事件</el-button
        >
    </div>
</template>

<script setup lang="ts" name="longpressDirect">
import { ElMessage } from 'element-plus';
const longpress = () => {
    ElMessage.success('长按事件触发成功 🎉🎉🎉');
};
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
