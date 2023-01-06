<template>
    <div>
        <el-card class="topRouterCard">
            <el-radio-group :model-value="props.activeName" @change="handleChange">
                <el-radio-button class="topRouterButton" size="default" label="operation">
                    {{ $t('logs.operation') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="login">
                    {{ $t('logs.login') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="system">
                    {{ $t('logs.system') }}
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
    activeName: 'operation',
});

const routerTo = (path: string) => {
    router.push({ path: path });
};

const handleChange = (val: string) => {
    switch (val) {
        case 'operation':
            routerTo('/logs');
            break;
        case 'login':
            routerTo('/logs/login');
            break;
        case 'system':
            routerTo('/logs/system');
            break;
    }
};
</script>
