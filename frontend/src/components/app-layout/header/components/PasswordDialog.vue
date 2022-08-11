<template>
    <el-dialog v-model="dialogVisible" title="修改密码" width="500px" draggable>
        <span>This is Password</span>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false">取消</el-button>
                <el-button type="primary" @click="dialogVisible = false">确认</el-button>
            </span>
        </template>
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
