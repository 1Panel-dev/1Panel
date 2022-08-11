<template>
    <el-dialog v-model="dialogVisible" title="个人资料" width="500px" draggable>
        <span>This is userInfo</span>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
const dialogVisible = ref(false);

// openDialog
const openDialog = () => {
    dialogVisible.value = true;
};

defineExpose({
    openDialog,
});
</script>
