<template>
    <!-- 列显隐设置 -->
    <el-drawer title="列设置" v-model="drawerVisible" size="400px">
        <div class="table-box">
            <el-table height="575" :data="colSetting" :border="true">
                <el-table-column prop="label" label="列名" />
                <el-table-column prop="name" label="显示" v-slot="scope">
                    <el-switch v-model="scope.row.isShow" @click="switchShow"></el-switch>
                </el-table-column>
                <template #empty>
                    <div class="table-empty">
                        <img src="@/assets/images/notData.png" alt="notData" />
                        <div>暂无数据</div>
                    </div>
                </template>
            </el-table>
        </div>
    </el-drawer>
</template>

<script setup lang="ts" name="colSetting">
import { ref, nextTick } from 'vue';
import { ColumnProps } from '@/components/ProTable/interface';

const props = defineProps<{
    colSetting: Partial<ColumnProps>[];
    tableRef: any;
}>();

const drawerVisible = ref<boolean>(false);
// 打开列设置
const openColSetting = (): void => {
    drawerVisible.value = true;
};

// 列显隐时重新布局 table（防止表格抖动,隐藏显示之后会出现横向滚动条,element-plus 内部问题，已经提了 issues）
const switchShow = () => {
    nextTick(() => {
        props.tableRef.doLayout();
    });
};

defineExpose({
    openColSetting,
});
</script>
