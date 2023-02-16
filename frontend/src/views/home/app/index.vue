<template>
    <div>
        <el-scrollbar max-height="500px">
            <div class="h-app-card" v-for="(app, index) in apps" :key="index">
                <el-row :gutter="10">
                    <el-col :span="5">
                        <div>
                            <el-avatar shape="square" :size="55" :src="'data:image/png;base64,' + app.icon" />
                        </div>
                    </el-col>
                    <el-col :span="15">
                        <div class="h-app-content">
                            <div>
                                <span class="h-app-title">{{ app.name }}</span>
                            </div>
                            <div class="h-app-desc">
                                <span>
                                    {{ app.shortDesc }}
                                </span>
                            </div>
                        </div>
                    </el-col>
                    <el-col :span="2">
                        <el-button
                            class="h-app-button"
                            type="primary"
                            plain
                            round
                            size="small"
                            @click="goInstall(app.key)"
                        >
                            {{ $t('app.install') }}
                        </el-button>
                    </el-col>
                </el-row>
            </div>
        </el-scrollbar>
    </div>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { Dashboard } from '@/api/interface/dashboard';
import { SearchApp } from '@/api/modules/app';
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
const router = useRouter();

let req = reactive({
    name: '',
    tags: [],
    page: 1,
    pageSize: 50,
    recommend: true,
});

const baseInfo = ref({
    haloID: 0,
    dateeaseID: 0,
    jumpserverID: 0,
    metersphereID: 0,
    kubeoperatorID: 0,
    kubepiID: 0,
});

let loading = ref(false);
let apps = ref<App.App[]>([]);

const acceptParams = (base: Dashboard.BaseInfo): void => {
    baseInfo.value.haloID = base.haloID;
    baseInfo.value.dateeaseID = base.dateeaseID;
    baseInfo.value.jumpserverID = base.jumpserverID;
    baseInfo.value.metersphereID = base.metersphereID;
    baseInfo.value.kubeoperatorID = base.kubeoperatorID;
    baseInfo.value.kubepiID = base.kubepiID;

    search(req);
};

const goInstall = (key: string) => {
    router.push({ name: 'AppDetail', params: { appKey: key } });
};

const search = async (req: App.AppReq) => {
    loading.value = true;
    await SearchApp(req)
        .then((res) => {
            apps.value = res.data.items;
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss">
.h-app-card {
    cursor: pointer;
    padding: 10px 15px;

    .h-app-content {
        .h-app-title {
            font-weight: 500;
            font-size: 15px;
        }

        .h-app-desc {
            span {
                font-weight: 400;
                font-size: 12px;
                color: #646a73;
            }
        }
    }
    .h-app-button {
        margin-top: 10px;
    }
    &:hover {
        background-color: rgba(0, 94, 235, 0.03);
    }
}

.h-app-divider {
    margin-top: 13px;
    border: 0;
    border-top: var(--panel-border);
}
</style>
