<template>
    <div>
        <div class="content-container__search">
            <el-card>
                <div>
                    <el-button
                        v-for="item in alertTabs"
                        :key="item.key"
                        class="tag-button"
                        :class="index === item.value ? '' : 'no-active'"
                        :type="index === item.value ? 'primary' : ''"
                        @click="changeTab(item.value)"
                    >
                        {{ $t(item.label) }}
                    </el-button>
                </div>
            </el-card>
        </div>
        <AlertDash v-if="index === ALERT_NOTICE_TAB.tasks" />
        <AlertLogs v-if="index === ALERT_NOTICE_TAB.logs" />
        <AlertSetting v-if="index === ALERT_NOTICE_TAB.config" />
    </div>
</template>
<script setup lang="ts">
import AlertDash from '@/views/setting/alert/dash/index.vue';
import AlertLogs from '@/views/setting/alert/log/index.vue';
import AlertSetting from '@/views/setting/alert/setting/index.vue';

const ALERT_NOTICE_TAB_STORAGE_KEY = 'alert-notice-tab';
const ALERT_NOTICE_TAB = {
    tasks: '0',
    logs: '1',
    config: '2',
} as const;
const alertTabs = [
    { key: 'tasks', value: ALERT_NOTICE_TAB.tasks, label: 'xpack.alert.list' },
    { key: 'logs', value: ALERT_NOTICE_TAB.logs, label: 'xpack.alert.logs' },
    { key: 'config', value: ALERT_NOTICE_TAB.config, label: 'commons.button.set' },
];
type AlertNoticeTabValue = (typeof ALERT_NOTICE_TAB)[keyof typeof ALERT_NOTICE_TAB];

const index = ref<AlertNoticeTabValue>(ALERT_NOTICE_TAB.tasks);

const isAlertNoticeTab = (value: string | null): value is AlertNoticeTabValue => {
    return alertTabs.some((item) => item.value === value);
};

const changeTab = (ind: AlertNoticeTabValue) => {
    index.value = ind;
    localStorage.setItem(ALERT_NOTICE_TAB_STORAGE_KEY, index.value);
};

onMounted(async () => {
    const tab = localStorage.getItem(ALERT_NOTICE_TAB_STORAGE_KEY);
    if (isAlertNoticeTab(tab)) {
        index.value = tab;
    }
});
</script>

<style scoped lang="scss">
.content-container__search {
    margin-top: 7px;

    :deep(.el-card) {
        --el-card-padding: 12px;
    }
}

.tag-button {
    &.no-active {
        background: none;
        border: none;
    }
}
</style>
