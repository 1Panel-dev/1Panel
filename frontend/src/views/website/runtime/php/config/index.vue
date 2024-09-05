<template>
    <DrawerPro v-model="open" :header="$t('runtime.runtime')" size="large" :resource="runtime.name" :back="handleClose">
        <template #content>
            <el-tabs tab-position="left" v-model="index">
                <el-tab-pane :label="$t('website.updateConfig')" name="0">
                    <Config :id="runtime.id" v-if="index == '0'"></Config>
                </el-tab-pane>
                <el-tab-pane :label="$t('php.disableFunction')" name="1">
                    <Function :id="runtime.id" v-if="index == '1'"></Function>
                </el-tab-pane>
                <el-tab-pane :label="$t('php.uploadMaxSize')" name="2">
                    <Upload :id="runtime.id" v-if="index == '2'"></Upload>
                </el-tab-pane>
                <el-tab-pane :label="'php-fpm'" name="3">
                    <PHP :id="runtime.id" v-if="index == '3'" :type="'fpm'"></PHP>
                </el-tab-pane>
                <el-tab-pane :label="'php'" name="4">
                    <PHP :id="runtime.id" v-if="index == '4'" :type="'php'"></PHP>
                </el-tab-pane>
            </el-tabs>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import Config from './config/index.vue';
import Function from './function/index.vue';
import Upload from './upload/index.vue';
import { Runtime } from '@/api/interface/runtime';
import PHP from './php-fpm/index.vue';

const index = ref('0');
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
