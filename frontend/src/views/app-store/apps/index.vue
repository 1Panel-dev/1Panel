<template>
    <el-card v-loading="loading">
        <el-row :gutter="5">
            <el-col :span="2">
                <el-button @click="sync" type="primary" plain="true">{{ $t('app.sync') }}</el-button>
            </el-col>
            <el-col :span="22">
                <div style="float: right">
                    <el-input
                        style="display: inline; margin-right: 5px"
                        v-model="req.name"
                        clearable
                        @clear="searchByName('')"
                    ></el-input>
                    <el-button
                        style="display: inline; margin-right: 5px"
                        v-model="req.name"
                        @click="searchByName(req.name)"
                    >
                        {{ '搜索' }}
                    </el-button>
                </div>
            </el-col>
        </el-row>
        <br />
        <el-row>
            <el-button style="margin-right: 5px" @click="changeTag('all')" type="primary" :plain="activeTag !== 'all'">
                {{ $t('app.all') }}
            </el-button>
            <div style="margin-right: 5px" :span="1" v-for="item in tags" :key="item.key">
                <el-button @click="changeTag(item.key)" type="primary" :plain="activeTag !== item.key">
                    {{ item.name }}
                </el-button>
            </div>
        </el-row>
        <el-row :gutter="5">
            <el-col v-for="(app, index) in apps" :key="index" :span="6">
                <div @click="getAppDetail(app.id)">
                    <el-card :body-style="{ padding: '0px' }" class="a-card">
                        <el-row :gutter="24">
                            <el-col :span="8">
                                <div class="icon">
                                    <el-image class="image" :src="'data:image/png;base64,' + app.icon"></el-image>
                                </div>
                            </el-col>
                            <el-col :span="16">
                                <div class="a-detail">
                                    <div class="d-name">
                                        <font size="3" style="font-weight: 700">{{ app.name }}</font>
                                    </div>
                                    <div class="d-description">
                                        <font size="1">
                                            <span>
                                                {{ app.shortDesc }}
                                            </span>
                                        </font>
                                    </div>
                                    <div class="d-tag">
                                        <el-tag v-for="(tag, ind) in app.tags" :key="ind" round :colr="getColor(ind)">
                                            {{ tag.name }}
                                        </el-tag>
                                    </div>
                                </div>
                            </el-col>
                        </el-row>
                    </el-card>
                </div>
            </el-col>
        </el-row>
    </el-card>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { onMounted, reactive, ref } from 'vue';
import router from '@/routers';
import { SearchApp, SyncApp } from '@/api/modules/app';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';

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
    let params: { [key: string]: any } = {
        id: id,
    };
    router.push({ name: 'AppDetail', params });
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
    height: 100px;
    margin-top: 10px;
    cursor: pointer;
    padding: 5px;

    .icon {
        width: 80px;
        height: 100%;
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
