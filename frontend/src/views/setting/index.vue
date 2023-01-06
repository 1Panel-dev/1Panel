<template>
    <div>
        <el-card class="topRouterCard">
            <el-radio-group :model-value="props.activeName" @change="handleChange">
                <el-radio-button class="topRouterButton" size="default" label="panel">
                    {{ $t('setting.panel') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="safe">
                    {{ $t('setting.safe') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="backupaccount">
                    {{ $t('setting.backupAccount') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="snapshot">
                    {{ $t('setting.snapshot') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="monitor">
                    {{ $t('menu.monitor') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="about">
                    {{ $t('setting.about') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import router from '@/routers';

interface MenuProps {
    activeName: string;
}
const props = withDefaults(defineProps<MenuProps>(), {
    activeName: 'all',
});

const routerTo = (path: string) => {
    router.push({ path: path });
};

const handleChange = (val: string) => {
    switch (val) {
        case 'panel':
            routerTo('/setting');
            break;
        case 'safe':
            routerTo('/setting/safe');
            break;
        case 'backupaccount':
            routerTo('/setting/backupaccount');
            break;
        case 'monitor':
            routerTo('/setting/monitor');
            break;
        case 'about':
            routerTo('/setting/about');
            break;
        case 'snapshot':
            routerTo('/setting/snapshot');
    }
};
</script>
