<template>
    <div class="content-box">
        <span class="text">批量上传数据 🍓🍇🍈🍉</span>
        <el-button type="primary" :icon="Upload" @click="batchAdd">批量上传数据</el-button>
        <ImportExcel ref="importRef"></ImportExcel>
    </div>
</template>

<script setup lang="ts" name="batchImport">
import { exportUserInfo, BatchAddUser } from '@/api/modules/user';
import ImportExcel from '@/components/ImportExcel/index.vue';
import { Upload } from '@element-plus/icons-vue';
import { ref } from 'vue';

interface DialogExpose {
    acceptParams: (params: any) => void;
}
const importRef = ref<DialogExpose>();
const batchAdd = () => {
    let params = {
        title: '数据',
        tempApi: exportUserInfo,
        importApi: BatchAddUser,
    };
    importRef.value!.acceptParams(params);
};
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
