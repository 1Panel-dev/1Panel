<template>
    <div>
        <el-card class="topRouterCard">
            <el-radio-group :model-value="props.activeName" @change="handleChange">
                <el-radio-button class="topRouterButton" size="default" label="container">
                    {{ $t('container.container') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="compose">
                    {{ $t('container.compose') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="image">
                    {{ $t('container.image') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="network">
                    {{ $t('container.network') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="volume">
                    {{ $t('container.volume') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="repo">
                    {{ $t('container.repo') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="template">
                    {{ $t('container.composeTemplate') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="setting">
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
