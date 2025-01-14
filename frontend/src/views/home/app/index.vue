<template>
    <div>
        <el-scrollbar height="525px" class="moz-height">
            <div class="h-app-card" v-for="(app, index) in apps" :key="index">
                <div class="flex justify-start items-center gap-2">
                    <div class="w-14">
                        <el-avatar shape="square" :size="55" :src="'data:image/png;base64,' + app.icon" />
                    </div>
                    <div class="flex-1 flex flex-col h-app-content">
                        <span class="h-app-title">{{ app.name }}</span>
                        <div class="h-app-desc">
                            <span>
                                {{ app.description }}
                            </span>
                        </div>
                    </div>
                    <div>
                        <el-button
                            class="h-app-button"
                            type="primary"
                            plain
                            round
                            size="small"
                            :disabled="app.limit == 1 && app.installed"
                            @click="goInstall(app.key, app.type)"
                        >
                            {{ $t('app.install') }}
                        </el-button>
                    </div>
                </div>
                <div class="h-app-divider" />
            </div>
        </el-scrollbar>
    </div>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
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

let loading = ref(false);
let apps = ref<App.AppDTO[]>([]);

const acceptParams = (): void => {
    search(req);
};

const goInstall = (key: string, type: string) => {
    switch (type) {
        case 'php':
        case 'node':
        case 'java':
        case 'go':
        case 'python':
        case 'dotnet':
            router.push({ path: '/websites/runtimes/' + type });
            break;
        default:
            router.push({ name: 'AppAll', query: { install: key } });
    }
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

<style lang="scss" scoped>
.h-app-card {
    cursor: pointer;
    padding: 10px 15px;
    margin-right: 10px;
    line-height: 18px;

    .h-app-content {
        padding-left: 15px;
        .h-app-title {
            font-weight: 500;
            font-size: 15px;
            color: var(--panel-text-color);
        }

        .h-app-desc {
            span {
                font-weight: 400;
                font-size: 12px;
                color: var(--el-text-color-regular);
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

/* FOR MOZILLA */
@-moz-document url-prefix() {
    .moz-height {
        height: 524px;
    }
}
</style>
