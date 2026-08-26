<template>
    <div v-if="visibleLinks.length" class="footer-navigation">
        <template v-for="(item, index) in visibleLinks" :key="item.key">
            <el-link type="primary" underline="never" @click="openLink(item.url)">
                <span class="font-normal">{{ $t(item.label) }}</span>
            </el-link>
            <el-divider v-if="index < visibleLinks.length - 1" direction="vertical" />
        </template>
        <el-divider class="footer-navigation__tail" direction="vertical" />
    </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { getEnterpriseFooterSetting } from '@/extensions/footer-setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
import {
    createDefaultFooterNavigationLinks,
    footerNavigationKeys,
    isSafeExternalUrl,
    mergeFooterNavigationLinks,
} from './model';
import type { FooterNavigationKey, FooterNavigationSetting } from './model';
import { FOOTER_NAVIGATION_REFRESH_EVENT } from './event';

const { docsUrl, isEE, isFxplay, isIntl } = useGlobalStore();
const setting = ref<FooterNavigationSetting | null>(null);
let loadGeneration = 0;

const defaults = computed(() => createDefaultFooterNavigationLinks(isIntl.value, docsUrl.value));
const links = computed(() => mergeFooterNavigationLinks(setting.value, defaults.value));
const labels: Record<FooterNavigationKey, string> = {
    learnMore: 'license.knowMorePro',
    forum: 'setting.forum',
    documentation: 'setting.doc2',
    project: 'setting.project',
};

const visibleLinks = computed(() => {
    return footerNavigationKeys
        .filter((key) => {
            if (isFxplay.value && key !== 'documentation') {
                return false;
            }
            return links.value[key].visible;
        })
        .map((key) => ({
            key,
            label: labels[key],
            url: links.value[key].url,
        }));
});

const loadSetting = async () => {
    const generation = ++loadGeneration;
    if (!isEE.value) {
        setting.value = null;
        return;
    }
    try {
        const res = await getEnterpriseFooterSetting(true);
        if (generation !== loadGeneration || !isEE.value) {
            return;
        }
        setting.value = res?.data || null;
    } catch {
        if (generation !== loadGeneration || !isEE.value) {
            return;
        }
        setting.value = null;
    }
};

const openLink = (url: string) => {
    if (!isSafeExternalUrl(url)) {
        return;
    }
    window.open(url, '_blank', 'noopener,noreferrer');
};

watch(isEE, loadSetting, { immediate: true });

onMounted(() => {
    window.addEventListener(FOOTER_NAVIGATION_REFRESH_EVENT, loadSetting);
});

onBeforeUnmount(() => {
    loadGeneration += 1;
    window.removeEventListener(FOOTER_NAVIGATION_REFRESH_EVENT, loadSetting);
});
</script>

<style scoped lang="scss">
.footer-navigation {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    row-gap: 8px;
}

:deep(.el-link__inner) {
    font-weight: 400;
}

@media (max-width: 767px) {
    .footer-navigation {
        column-gap: 12px;
        justify-content: center;
    }

    .footer-navigation :deep(.el-divider--vertical) {
        display: none;
    }
}
</style>
