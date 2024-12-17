<template>
    <div class="footer" :style="{ height: mobile ? '108px' : '48px' }">
        <div class="flex w-full flex-col gap-4 md:justify-between md:flex-row">
            <div class="flex flex-wrap gap-4">
                <a href="https://fit2cloud.com/" target="_blank">Copyright © 2014-2024 {{ $t('commons.fit2cloud') }}</a>
            </div>
            <div class="flex flex-row gap-2 md:flex-col lg:flex-row">
                <SystemUpgrade :footer="true" />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import SystemUpgrade from '@/components/system-upgrade/index.vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});
</script>

<style scoped lang="scss">
.footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 48px;
    background: var(--panel-footer-bg);
    border-top: 1px solid var(--panel-footer-border);
    box-sizing: border-box;
    padding: 10px 20px;
    a {
        font-size: 12px;
        color: #858585;
        text-decoration: none;
        letter-spacing: 0.5px;
    }
    span {
        font-size: 12px;
        color: #858585;
        text-decoration: none;
        letter-spacing: 0.5px;
    }
}
</style>
