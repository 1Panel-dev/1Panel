<template>
    <div>
        <el-card class="topRouterCard">
            <el-radio-group :model-value="props.activeName" @change="handleChange">
                <el-radio-button class="topRouterButton" size="default" label="mysql">Mysql</el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="redis">Redis</el-radio-button>
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
