<template>
    <div>
        <el-tabs tab-position="left" v-model="tabIndex" ref="tabsRef">
            <el-tab-pane :label="$t('nginx.defaultHttps')" name="0">
                <Https v-if="tabIndex == '0'"></Https>
            </el-tab-pane>
        </el-tabs>
    </div>
</template>

<script lang="ts" setup>
import Https from './https/index.vue';

const tabIndex = ref('0');
</script>
