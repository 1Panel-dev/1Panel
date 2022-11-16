<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="$t('website.currentSSL')">
            <Current :id="id" v-if="index == '0'"></Current>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.dnsAccount')">
            <Account :id="id" v-if="index == '1'"></Account>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.applySSL')">
            <SSL :id="id" v-if="index == '2'"></SSL>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import Current from './current/index.vue';
import Account from './account/index.vue';
import SSL from './ssl/index.vue';

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
