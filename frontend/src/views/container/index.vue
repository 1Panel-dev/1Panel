<template>
    <div>
        <el-card class="topCard">
            <el-radio-group :model-value="props.activeName" @change="handleChange">
                <el-radio-button class="topButton" size="large" label="container">
                    {{ $t('container.container') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="compose">
                    {{ $t('container.compose') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="image">
                    {{ $t('container.image') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="network">
                    {{ $t('container.network') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="volume">
                    {{ $t('container.volume') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="repo">
                    {{ $t('container.repo') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="template">
                    {{ $t('container.composeTemplate') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="setting">
                    {{ $t('container.setting') }}
                </el-radio-button>
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
    activeName: 'container',
});

const routerTo = (path: string) => {
    router.push({ path: path });
};

const handleChange = (val: string) => {
    switch (val) {
        case 'container':
            routerTo('/containers');
            break;
        case 'compose':
            routerTo('/containers/compose');
            break;
        case 'image':
            routerTo('/containers/image');
            break;
        case 'network':
            routerTo('/containers/network');
            break;
        case 'volume':
            routerTo('/containers/volume');
            break;
        case 'repo':
            routerTo('/containers/repo');
            break;
        case 'template':
            routerTo('/containers/template');
            break;
        case 'setting':
            routerTo('/containers/setting');
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
