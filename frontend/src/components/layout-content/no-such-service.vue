<template>
    <div>
        <LayoutContent :divider="true">
            <template #main>
                <div class="app-warn">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <span>{{ $t('cronjob.library.noSuchApp', [prop.name]) }}</span>
                        <span @click="toDoc" class="flex items-center justify-center gap-0.5">
                            <el-icon><Position /></el-icon>
                            {{ $t('firewall.quickJump') }}
                        </span>
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
import router from '@/routers';

const prop = defineProps({
    name: String,
});

const toDoc = () => {
    router.push({ name: 'Library', query: { uncached: 'true' } });
};
</script>
