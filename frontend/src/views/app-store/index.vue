<template>
    <div>
        <el-card class="topCard">
            <el-radio-group v-model="activeName" @change="handleChange">
                <el-radio-button class="topButton" size="large" label="all">
                    {{ $t('app.all') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="installed">
                    {{ $t('app.installed') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>
        <br />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

const activeName = ref('all');

const routerTo = (path: string) => {
    router.push({ path: path });
};

const handleChange = (val: string) => {
    switch (val) {
        case 'all':
            routerTo('/apps/all');
            break;
        case 'installed':
            routerTo('/apps/installed');
            break;
    }
};

onMounted(() => {
    const path = router.currentRoute.value.path;
    if (path === '/apps/all') {
        activeName.value = 'all';
    }
    if (path === '/apps/installed') {
        activeName.value = 'installed';
    }
    if (path === '/apps') {
        routerTo('/apps/all');
    }
});
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
