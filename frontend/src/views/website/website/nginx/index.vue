<template>
    <LayoutContent :header="$t('website.nginxConfig')" :reload="true">
        <el-collapse v-model="activeName" accordion>
            <el-collapse-item :title="$t('website.source')" name="1">
                <Source></Source>
            </el-collapse-item>
        </el-collapse>
    </LayoutContent>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import Source from './source/index.vue';
import { ref } from 'vue';

let activeName = ref('1');
</script>
