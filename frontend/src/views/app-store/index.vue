<template>
    <div>
        <el-row :gutter="20">
            <el-col :span="24">
                <div style="margin-bottom: 10px">
                    <el-check-tag :checked="activeName === 'all'" @click="routerTo('/apps/all')">
                        {{ $t('app.all') }}
                    </el-check-tag>
                    <el-check-tag :checked="activeName === 'installed'" @click="routerTo('/apps/installed')">
                        {{ $t('app.installed') }}
                    </el-check-tag>
                </div>
            </el-col>
        </el-row>
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

<style lang="scss"></style>
