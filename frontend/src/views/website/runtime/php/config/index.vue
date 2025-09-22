<template>
    <DrawerPro
        v-model="open"
        :header="$t('runtime.runtime')"
        size="large"
        :resource="runtime.name"
        @close="handleClose"
    >
        <template #content>
            <el-tabs tab-position="left" v-model="index">
                <el-tab-pane :label="$t('php.containerConfig')" name="6">
                    <Container :id="runtime.id" v-if="index == '6'"></Container>
                </el-tab-pane>
                <el-tab-pane :label="$t('website.updateConfig')" name="0">
                    <Config :id="runtime.id" v-if="index == '0'"></Config>
                </el-tab-pane>
                <el-tab-pane :label="$t('php.disableFunction')" name="1">
                    <Function :id="runtime.id" v-if="index == '1'"></Function>
                </el-tab-pane>
                <el-tab-pane :label="$t('php.uploadMaxSize')" name="2">
                    <Upload :id="runtime.id" v-if="index == '2'"></Upload>
                </el-tab-pane>
                <el-tab-pane :label="$t('cronjob.timeout')" name="8">
                    <Timeout :id="runtime.id" v-if="index == '8'"></Timeout>
                </el-tab-pane>
                <el-tab-pane :label="$t('website.nginxPer')" name="5">
                    <Performance :id="runtime.id" v-if="index == '5'"></Performance>
                </el-tab-pane>
                <el-tab-pane :label="$t('runtime.loadStatus')" name="7">
                    <FpmStatus :id="runtime.id" v-if="index == '7'"></FpmStatus>
                </el-tab-pane>
                <el-tab-pane :label="$t('website.source')" name="4">
                    <PHP :id="runtime.id" v-if="index == '4'" :type="'php'"></PHP>
                </el-tab-pane>
                <el-tab-pane :label="'FPM ' + $t('website.source')" name="3">
                    <PHP :id="runtime.id" v-if="index == '3'" :type="'fpm'"></PHP>
                </el-tab-pane>
            </el-tabs>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import Config from './config/index.vue';
import Function from './function/index.vue';
import Upload from './upload/index.vue';
import PHP from './php-fpm/index.vue';
import Performance from './performance/index.vue';
import Container from './container/index.vue';
import FpmStatus from './fpm-status/index.vue';
import Timeout from './timeout/index.vue';

const index = ref('6');
const open = ref(false);
const runtime = ref({
    name: '',
    id: 0,
});

const handleClose = () => {
    open.value = false;
};

const acceptParams = async (req: Runtime.Runtime) => {
    runtime.value = req;
    open.value = true;
};

onMounted(() => {});

defineExpose({
    acceptParams,
});
</script>
