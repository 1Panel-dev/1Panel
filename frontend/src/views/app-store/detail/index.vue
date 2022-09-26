<template>
    <LayoutContent :header="$t('app.detail')" :back-name="'App'">
        <div class="brief">
            <el-row :gutter="20">
                <el-col :span="4">
                    <div class="icon">
                        <el-image class="image" :src="'data:image/png;base64,' + app.icon"></el-image>
                    </div>
                </el-col>
                <el-col :span="20">
                    <div class="a-detail">
                        <div class="a-name">
                            <font size="5" style="font-weight: 800">{{ app.name }}</font>
                        </div>
                        <div class="a-description">
                            <span>
                                <font>
                                    {{ app.shortDesc }}
                                </font>
                            </span>
                        </div>
                        <br />
                        <el-descriptions :column="1">
                            <el-descriptions-item :label="$t('app.version')">
                                <el-select v-model="version" @change="getDetail(version)">
                                    <el-option v-for="(v, index) in app.versions" :key="index" :value="v" :label="v">
                                        {{ v }}
                                    </el-option>
                                </el-select>
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('app.source')">
                                <el-link @click="toLink(app.source)">
                                    <el-icon><Link /></el-icon>
                                </el-link>
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('app.author')">{{ app.author }}</el-descriptions-item>
                        </el-descriptions>
                        <div>
                            <el-button @click="openInstall" type="primary">{{ $t('app.install') }}</el-button>
                        </div>
                    </div>
                </el-col>
            </el-row>
        </div>
        <el-divider border-style="double" />
        <div class="detail" v-loading="loadingDetail">
            <v-md-preview :text="appDetail.readme"></v-md-preview>
        </div>
        <Install ref="installRef"></Install>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { GetApp, GetAppDetail } from '@/api/modules/app';
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, ref } from 'vue';
import Install from './install.vue';

interface OperateProps {
    id: number;
}
const props = withDefaults(defineProps<OperateProps>(), {
    id: 0,
});
let app = ref<any>({});
let appDetail = ref<any>({});
let version = ref('');
let loadingDetail = ref(false);
const installRef = ref();

const getApp = () => {
    GetApp(props.id).then((res) => {
        app.value = res.data;
        version.value = app.value.versions[0];
        getDetail(version.value);
    });
};

const getDetail = (version: string) => {
    loadingDetail.value = true;
    GetAppDetail(props.id, version)
        .then((res) => {
            appDetail.value = res.data;
        })
        .finally(() => {
            loadingDetail.value = false;
        });
};

const toLink = (link: string) => {
    window.open(link, '_blank');
};

const openInstall = () => {
    let params = {
        params: appDetail.value.params,
        appDetailId: appDetail.value.id,
    };
    installRef.value.acceptParams(params);
};

onMounted(() => {
    getApp();
});
</script>

<style lang="scss">
.brief {
    height: 30vh;
    .icon {
        .image {
            width: auto;
            height: 20vh;
        }
    }
}
</style>
