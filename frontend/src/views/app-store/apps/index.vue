<template>
    <LayoutContent v-loading="loading" v-if="!showDetail" :title="$t('website.website')">
        <template #toolbar>
            <el-row :gutter="5">
                <el-col :span="20">
                    <div>
                        <el-button @click="changeTag('all')" type="primary" :plain="activeTag !== 'all'">
                            {{ $t('app.all') }}
                        </el-button>
                        <div v-for="item in tags" :key="item.key" style="display: inline">
                            <el-button
                                class="tag-button"
                                @click="changeTag(item.key)"
                                type="primary"
                                :plain="activeTag !== item.key"
                            >
                                {{ item.name }}
                            </el-button>
                        </div>
                    </div>
                </el-col>
                <el-col :span="4">
                    <div style="float: right">
                        <el-input
                            class="table-button"
                            v-model="req.name"
                            clearable
                            @clear="searchByName('')"
                            suffix-icon="Search"
                            @keyup.enter="searchByName(req.name)"
                            @blur="searchByName(req.name)"
                            :placeholder="$t('commons.button.search')"
                        ></el-input>
                    </div>
                </el-col>
            </el-row>
        </template>
        <template #rightButton>
            <el-button @click="sync" type="text" :plain="true">{{ $t('app.syncAppList') }}</el-button>
        </template>
        <template #main>
            <div class="divider"></div>
            <el-row :gutter="5">
                <el-col v-for="(app, index) in apps" :key="index" :span="8">
                    <div class="a-card">
                        <el-row :gutter="24">
                            <el-col :span="5">
                                <div class="icon">
                                    <el-avatar shape="square" :size="60" :src="'data:image/png;base64,' + app.icon" />
                                </div>
                            </el-col>
                            <el-col :span="19">
                                <div class="a-detail">
                                    <div class="d-name">
                                        <span class="name">{{ app.name }}</span>
                                        <el-button class="h-button" round @click="getAppDetail(app.id)">安装</el-button>
                                    </div>
                                    <div class="d-description">
                                        <span class="description">
                                            {{ app.shortDesc }}
                                        </span>
                                    </div>
                                    <div class="d-tag" style="margin-top: 5px">
                                        <el-tag v-for="(tag, ind) in app.tags" :key="ind" :colr="getColor(ind)">
                                            {{ tag.name }}
                                        </el-tag>
                                    </div>
                                    <div class="divider"></div>
                                </div>
                            </el-col>
                        </el-row>
                    </div>
                </el-col>
            </el-row>
        </template>
    </LayoutContent>
    <Detail v-if="showDetail" :id="appId"></Detail>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import { App } from '@/api/interface/app';
import { onMounted, reactive, ref } from 'vue';
import { SearchApp, SyncApp } from '@/api/modules/app';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';
import Detail from '../detail/index.vue';

let req = reactive({
    name: '',
    tags: [],
    page: 1,
    pageSize: 50,
});

let apps = ref<App.App[]>([]);
let tags = ref<App.Tag[]>([]);
const colorArr = ['#6495ED', '#54FF9F', '#BEBEBE', '#FFF68F', '#FFFF00', '#8B0000'];
let loading = ref(false);
let activeTag = ref('all');
let showDetail = ref(false);
let appId = ref(0);

const getColor = (index: number) => {
    return colorArr[index];
};

const search = async (req: App.AppReq) => {
    await SearchApp(req).then((res) => {
        apps.value = res.data.items;
        tags.value = res.data.tags;
    });
};

const getAppDetail = (id: number) => {
    showDetail.value = true;
    appId.value = id;
};

const sync = () => {
    loading.value = true;
    SyncApp()
        .then(() => {
            ElMessage.success(i18n.global.t('app.syncSuccess'));
            search(req);
        })
        .finally(() => {
            loading.value = false;
        });
};

const changeTag = (key: string) => {
    req.tags = [];
    activeTag.value = key;
    if (key !== 'all') {
        req.tags = [key];
    }
    search(req);
};

const searchByName = (name: string) => {
    req.name = name;
    search(req);
};

onMounted(() => {
    search(req);
});
</script>

<style lang="scss">
.header {
    padding-bottom: 10px;
}

.a-card {
    height: 120px;
    margin-top: 10px;
    cursor: pointer;
    padding: 5px;

    .icon {
        margin-left: 10px;
        width: 80px;
        height: 80%;
        padding: 5px;
        display: flex;
        align-items: center;
        text-align: center;
        .image {
            width: 80%;
            height: 80%;
            margin: 0 auto;
        }
    }

    .a-detail {
        margin-top: 10px;
        height: 100%;
        width: 100%;

        .d-name {
            height: 20%;
            .name {
                font-weight: 500;
                font-size: 16px;
                color: #1f2329;
            }
            .h-button {
                float: right;
            }
        }

        .d-description {
            margin-top: 10px;
            overflow: hidden;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
            .description {
                font-size: 14px;
                color: #646a73;
            }
        }
    }
}

.a-card:hover {
    background-color: rgba(0, 94, 235, 0.03);
}

.table-button {
    display: inline;
    margin-right: 5px;
}

.tag-button {
    margin-left: 10px;
}

.divider {
    margin-top: 5px;
    border: 0;
    border-top: 1px solid #ccc;
}
</style>
