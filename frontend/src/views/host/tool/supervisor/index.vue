<template>
    <div>
        <ToolRouter />
        <LayoutContent :title="$t('tool.supervisor.list')" v-loading="loading">
            <template #app>
                <SuperVisorStatus @setting="setting" v-model:loading="loading" @is-exist="isExist" />
            </template>
            <template v-if="isExistSuperVisor && !setSuperVisor" #toolbar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('commons.button.create') + $t('tool.supervisor.list') }}
                </el-button>
            </template>
            <template #main v-if="isExistSuperVisor && !setSuperVisor">
                <ComplexTable></ComplexTable>
            </template>
            <ConfigSuperVisor v-if="setSuperVisor" />
        </LayoutContent>
        <Create ref="createRef"></Create>
    </div>
</template>

<script setup lang="ts">
import ToolRouter from '@/views/host/tool/index.vue';
import SuperVisorStatus from './status/index.vue';
import { ref } from '@vue/runtime-core';
import ConfigSuperVisor from './config/index.vue';
import { onMounted } from 'vue';
import Create from './create/index.vue';

const loading = ref(false);
const setSuperVisor = ref(false);
const isExistSuperVisor = ref(false);
const createRef = ref();

const setting = () => {
    setSuperVisor.value = true;
};

const isExist = (isExist: boolean) => {
    isExistSuperVisor.value = isExist;
};

const openCreate = () => {
    createRef.value.acceptParams();
};

onMounted(() => {});
</script>
