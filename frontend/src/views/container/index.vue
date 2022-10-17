<template>
    <div>
        <el-card class="topCard">
            <el-radio-group v-model="activeNames">
                <el-radio-button class="topButton" size="large" label="container">
                    {{ $t('container.container') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="image">
                    {{ $t('container.image') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="network">
                    {{ $t('container.network') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="volume">
                    {{ $t('container.volume') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="repo">
                    {{ $t('container.repo') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="compose">
                    {{ $t('container.compose') }}
                </el-radio-button>
                <el-radio-button class="topButton" size="large" label="composeTemplate">
                    {{ $t('container.composeTemplate') }}
                </el-radio-button>
            </el-radio-group>
        </el-card>
        <Container v-if="activeNames === 'container'" />
        <Repo v-if="activeNames === 'repo'" />
        <Image v-if="activeNames === 'image'" />
        <Network v-if="activeNames === 'network'" />
        <Volume v-if="activeNames === 'volume'" />
        <ComposeTemplate v-if="activeNames === 'composeTemplate'" />
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import Container from '@/views/container/container/index.vue';
import Repo from '@/views/container/repo/index.vue';
import Image from '@/views/container/image/index.vue';
import Network from '@/views/container/network/index.vue';
import Volume from '@/views/container/volume/index.vue';
import ComposeTemplate from '@/views/container/compose/template/index.vue';

const activeNames = ref('container');
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
