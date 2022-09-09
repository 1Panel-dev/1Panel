<template>
    <el-dialog v-model="dialogVisible" :title="$t('commons.header.changePassword')" width="500px" draggable>
        <span>This is Password</span>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="dialogVisible = false">{{ $t('commons.button.confirm') }}</el-button>
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
