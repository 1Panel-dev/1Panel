<template>
    <div>
        <Submenu activeName="redis" />

        <AppStatus :app-key="'redis'" style="margin-top: 20px" @setting="onSetting"></AppStatus>
        <Setting ref="settingRef"></Setting>

        <Terminal v-if="!isOnSetting" style="margin-top: 5px" ref="terminalRef"></Terminal>
    </div>
</template>

<script lang="ts" setup>
import Submenu from '@/views/database/index.vue';
import Setting from '@/views/database/redis/setting/index.vue';
import Terminal from '@/views/database/redis/terminal/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import { onMounted, ref } from 'vue';

const terminalRef = ref();
const settingRef = ref();
const isOnSetting = ref(false);

const onSetting = async () => {
    isOnSetting.value = true;
    terminalRef.value.onClose();
    settingRef.value!.acceptParams();
};

onMounted(() => {
    terminalRef.value.acceptParams();
});
</script>
