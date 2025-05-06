<template>
    <div v-if="row.port != ''">
        <span v-for="(port, index) in row.port.split(',')" :key="index">
            <el-button link @click="jump(port, 'http')">
                {{ port }}
                <el-icon class="el-icon--right"><Promotion /></el-icon>
            </el-button>
        </span>
    </div>
</template>

<script setup>
import { Promotion } from '@element-plus/icons-vue';
defineProps({
    row: {
        type: Object,
        required: true,
    },
    jump: {
        type: Function,
        required: true,
    },
});
</script>
