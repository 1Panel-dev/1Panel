<template>
    <LayoutContent>
        <el-row :gutter="20">
            <el-col :span="24">
                <div style="margin-bottom: 10px">
                    <el-radio-group v-model="activeName">
                        <el-radio-button label="all">
                            {{ $t('app.all') }}
                        </el-radio-button>
                        <el-radio-button label="installed">
                            {{ $t('app.installed') }}
                        </el-radio-button>
                    </el-radio-group>
                    <div style="float: right">
                        <el-button @click="sync">{{ $t('app.sync') }}</el-button>
                    </div>
                </div>
            </el-col>
        </el-row>
        <Apps v-if="activeName === 'all'"></Apps>
        <Installed v-if="activeName === 'installed'"></Installed>
    </LayoutContent>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import { ref } from 'vue';
import { SyncApp } from '@/api/modules/app';
import Apps from './apps/index.vue';
import Installed from './installed/index.vue';
const activeName = ref('all');

const sync = () => {
    SyncApp().then((res) => {
        console.log(res);
    });
};
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
