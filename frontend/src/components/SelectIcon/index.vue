<template>
    <div class="icon-box">
        <el-input
            v-model="iconValue"
            placeholder="请选择图标"
            @focus="openDialog"
            readonly
            ref="inputRef"
        >
            <template #append>
                <el-button :icon="customIcons[iconValue]" />
            </template>
        </el-input>
        <el-dialog
            v-model="dialogVisible"
            title="请选择图标"
            top="50px"
            width="1280px"
        >
            <div
                v-for="(item, index) in Icons"
                :key="index"
                class="icon-item"
                @click="selectIcon(item)"
            >
                <component :is="item"></component>
                <span>{{ item.name }}</span>
            </div>
        </el-dialog>
    </div>
</template>

<script setup lang="ts" name="selectIcon">
import { ref } from 'vue';
import * as Icons from '@element-plus/icons-vue';

// 接收参数
defineProps<{ iconValue: string }>();

const customIcons: { [key: string]: any } = Icons;
const dialogVisible = ref(false);

// 打开 dialog
const openDialog = (e: any) => {
    // 直接让文本框失去焦点，不然会出现显示bug
    e.srcElement.blur();
    dialogVisible.value = true;
};

const emit = defineEmits(['update:iconValue']);

// 选择图标(触发更新父组件数据)
const selectIcon = (item: any) => {
    dialogVisible.value = false;
    emit('update:iconValue', item.name);
};
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
