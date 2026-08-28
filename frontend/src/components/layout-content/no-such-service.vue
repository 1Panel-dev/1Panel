<template>
    <div>
        <LayoutContent :divider="true">
            <template #main>
                <div class="app-warn">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <slot>
                            <span>{{ $t('cronjob.library.noSuchApp', [prop.name]) }}</span>
                            <span @click="toDoc" class="flex items-center justify-center gap-0.5">
                                <el-icon><Position /></el-icon>
                                {{ $t('firewall.quickJump') }}
                            </span>
                        </slot>
                    </div>
                    <div>
                        <img src="@/assets/images/no_app.svg" />
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { routerToNameWithQuery } from '@/utils/router';

const prop = defineProps({
    name: String,
});

const toDoc = () => {
    routerToNameWithQuery('Library', { uncached: 'true' });
};
</script>

<style scoped>
:deep(.firewall-backend-link) {
    appearance: none;
    margin: 0;
    margin-inline: 4px;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--el-color-primary);
    cursor: pointer;
    font-family: inherit;
    font-size: inherit;
    font-weight: inherit;
    line-height: inherit;
    text-decoration: underline;
    text-underline-offset: 3px;
    vertical-align: baseline;
}

:deep(.firewall-backend-link:hover) {
    color: var(--el-color-primary-light-3);
}

:deep(.firewall-backend-link:active) {
    color: var(--el-color-primary-dark-2);
}

:deep(.firewall-backend-link:focus-visible) {
    border-radius: 2px;
    outline: 2px solid var(--el-color-primary-light-5);
    outline-offset: 2px;
}
</style>
