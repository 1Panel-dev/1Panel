<template>
    <div>
        <el-card class="router_card p-2 sm:p-3 mt-2">
            <div class="flex w-full justify-start items-center">
                <el-button type="primary" :plain="index !== '0'" @click="changeTab('0')">
                    {{ $t('xpack.alert.list') }}
                </el-button>
                <el-button type="primary" :plain="index !== '1'" @click="changeTab('1')">
                    {{ $t('xpack.alert.logs') }}
                </el-button>
                <el-button type="primary" :plain="index !== '2'" @click="changeTab('2')">
                    {{ $t('commons.button.set') }}
                </el-button>
            </div>
        </el-card>
        <AlertDash v-if="index == '0'" />
        <AlertLogs v-if="index == '1'" />
        <AlertSetting v-if="index == '2'" />
    </div>
</template>
<script setup lang="ts">
import AlertDash from '@/views/setting/alert/dash/index.vue';
import AlertLogs from '@/views/setting/alert/log/index.vue';
import AlertSetting from '@/views/setting/alert/setting/index.vue';

const index = ref('0');

const changeTab = (ind: string) => {
    index.value = ind;
    localStorage.setItem('alert-notice-tab', index.value);
};

onMounted(async () => {
    const tab = localStorage.getItem('alert-notice-tab');
    if (tab) {
        index.value = tab;
    }
});
</script>
