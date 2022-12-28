<template>
    <div>
        <el-card class="topCard">
            <el-radio-group :model-value="props.activeName" @change="handleChange">
                <el-radio-button class="topButton" size="large" label="mysql">Mysql</el-radio-button>
                <el-radio-button class="topButton" size="large" label="redis">Redis</el-radio-button>
            </el-radio-group>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
const router = useRouter();
interface MenuProps {
    activeName: string;
}
const props = withDefaults(defineProps<MenuProps>(), {
    activeName: 'mysql',
});

const routerTo = (path: string) => {
    router.push({ path: path });
};

const handleChange = (val: string) => {
    switch (val) {
        case 'mysql':
            routerTo('/databases');
            break;
        case 'redis':
            routerTo('/databases/redis');
            break;
    }
};
</script>

<style>
.topCard {
    --el-card-border-color: var(--el-border-color-light);
    --el-card-border-radius: 4px;
    --el-card-padding: 0px;
    --el-card-bg-color: var(--el-fill-color-blank);
}
.topButton .el-radio-button__inner {
    display: inline-block;
    line-height: 1;
    white-space: nowrap;
    vertical-align: middle;
    background: var(--el-button-bg-color, var(--el-fill-color-blank));
    border: 0;
    font-weight: 350;
    border-left: 0;
    color: var(--el-button-text-color, var(--el-text-color-regular));
    text-align: center;
    box-sizing: border-box;
    outline: 0;
    margin: 0;
    position: relative;
    cursor: pointer;
    transition: var(--el-transition-all);
    -webkit-user-select: none;
    user-select: none;
    padding: 8px 15px;
    font-size: var(--el-font-size-base);
    border-radius: 0;
}
</style>
