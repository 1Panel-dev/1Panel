<template>
    <div>
        <div class="content-container__search">
            <el-card>
                <div>
                    <el-button
                        class="tag-button"
                        :class="index === '0' ? '' : 'no-active'"
                        :type="index === '0' ? 'primary' : ''"
                        @click="changeTab('0')"
                    >
                        {{ $t('xpack.alert.list') }}
                    </el-button>
                    <el-button
                        class="tag-button"
                        :class="index === '1' ? '' : 'no-active'"
                        :type="index === '1' ? 'primary' : ''"
                        @click="changeTab('1')"
                    >
                        {{ $t('xpack.alert.logs') }}
                    </el-button>
                    <el-button
                        class="tag-button"
                        :class="index === '2' ? '' : 'no-active'"
                        :type="index === '2' ? 'primary' : ''"
                        @click="changeTab('2')"
                    >
                        {{ $t('commons.button.set') }}
                    </el-button>
                </div>
            </el-card>
        </div>
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
