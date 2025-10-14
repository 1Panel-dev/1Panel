<template>
    <LayoutContent :title="$t('tool.supervisor.config')" :reload="true">
        <template #leftToolBar>
            <el-button type="primary" :plain="activeName !== '1'" @click="changeTab('1')">
                {{ $t('nginx.configResource') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '2'" @click="changeTab('2')">
                {{ $t('commons.button.log') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '3'" @click="changeTab('3')">
                {{ $t('commons.button.init') }}
            </el-button>
        </template>
        <template #main>
            <Source v-if="activeName === '1'"></Source>
            <div v-if="activeName === '2'">
                <LogFile
                    :config="{ id: 0, type: 'supervisord', name: 'supervisor', colorMode: 'container' }"
                    ref="logRef"
                ></LogFile>
            </div>
            <Basic v-if="activeName === '3'"></Basic>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import Source from './source/index.vue';
import LogFile from '@/components/log/file/index.vue';
import Basic from './basic/index.vue';

const activeName = ref('1');

const changeTab = (index: string) => {
    activeName.value = index;
};
</script>
