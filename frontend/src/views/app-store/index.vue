<template>
    <div>
        <el-card class="topRouterCard">
            <el-radio-group v-model="activeName" @change="handleChange">
                <el-radio-button class="topRouterButton" size="default" label="all">
                    {{ $t('app.all') }}
                </el-radio-button>
                <el-radio-button class="topRouterButton" size="default" label="installed">
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
