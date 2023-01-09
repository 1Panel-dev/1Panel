<template>
    <el-tabs tab-position="left" v-model="index">
        <el-tab-pane :label="$t('website.accessLog')" name="0">
            <LogFile :id="id" :log-type="'access.log'" v-if="index == '0'"></LogFile>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.errLog')" name="1">
            <LogFile :id="id" :log-type="'error.log'" v-if="index == '1'"></LogFile>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import LogFile from './log-fiile/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});

let index = ref('0');
</script>
