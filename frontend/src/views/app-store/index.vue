<template>
    <div>
        <el-row :gutter="20">
            <el-col :span="24">
                <div style="margin-bottom: 10px">
                    <el-radio-group v-model="activeName">
                        <el-radio-button label="all" @click="routerTo('/apps/all')">
                            {{ $t('app.all') }}
                        </el-radio-button>
                        <el-radio-button label="installed" @click="routerTo('/apps/installed')">
                            {{ $t('app.installed') }}
                        </el-radio-button>
                    </el-radio-group>
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

// const sync = () => {
//     SyncApp().then((res) => {
//         console.log(res);
//     });
// };

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

<style lang="scss">
.header {
    padding-bottom: 10px;
}

.a-card {
    height: 100px;
    margin-top: 10px;
    cursor: pointer;
    padding: 1px;

    .icon {
        width: 100%;
        height: 80%;
        padding: 10%;
        margin-top: 5px;
        .image {
            width: auto;
            height: auto;
        }
    }

    .a-detail {
        margin-top: 10px;
        height: 100%;
        width: 100%;

        .d-name {
            height: 20%;
        }

        .d-description {
            overflow: hidden;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
        }
    }
}

.a-card:hover {
    transform: scale(1.1);
}
</style>
