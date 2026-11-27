<template>
    <div class="app-status">
        <el-card>
            <div>
                <el-select
                    class="website-select"
                    :model-value="props.id"
                    :aria-label="$t('menu.website')"
                    filterable
                    remote
                    remote-show-suffix
                    :remote-method="search"
                    :loading="loading && page === 0"
                    @end-reached="loadMore"
                    @change="(id) => emit('select', id)"
                >
                    <template #label>{{ props.primaryDomain }}</template>
                    <el-option v-for="site in options" :key="site.id" :value="site.id" :label="site.primaryDomain" />
                </el-select>
                <span class="status-content">
                    <Status class="span-font" :key="props.status" :status="props.status"></Status>
                </span>
                <span class="status-content">
                    <el-tag type="info">
                        {{ $t('website.expireDate') }}:
                        <span v-if="isEver(props.expireDate)">
                            {{ $t('website.neverExpire') }}
                        </span>
                        <span v-else>{{ dateFormatSimple(props.expireDate) }}</span>
                    </el-tag>
                </span>
            </div>
        </el-card>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { searchWebsites } from '@/api/modules/website';
import type { Website } from '@/api/interface/website';
import Status from '@/components/status/index.vue';
import { dateFormatSimple } from '@/utils/date';
const props = defineProps({
    id: { type: Number, required: true },
    primaryDomain: {
        type: String,
        default: '',
    },
    status: {
        type: String,
        default: '',
    },
    expireDate: {
        type: String,
        default: '',
    },
});

const emit = defineEmits<{ select: [id: number] }>();
const options = ref<Website.WebsiteRes[]>([]);
const loading = ref(false);
const page = ref(0);
let total = 0;
let query = '';
let revision = 0;

const fetchPage = async () => {
    const request = revision;
    const nextPage = page.value + 1;
    loading.value = true;
    try {
        const { data } = await searchWebsites({
            page: nextPage,
            pageSize: 20,
            name: query,
            orderBy: 'favorite',
            order: 'descending',
            websiteGroupId: 0,
        });
        if (request !== revision) return;
        options.value = options.value.concat(data.items ?? []);
        total = data.total;
        page.value = nextPage;
    } finally {
        if (request === revision) loading.value = false;
    }
};

const search = (name: string) => {
    if (query === name && (loading.value || page.value > 0)) return;
    revision++;
    query = name;
    page.value = 0;
    total = 0;
    options.value = [];
    return fetchPage();
};

const loadMore = (direction: string) => {
    if (direction === 'bottom' && !loading.value && options.value.length < total) return fetchPage();
};

const isEver = (time: string) => {
    const expireDate = new Date(time);
    return expireDate > new Date('9999-12-30');
};
</script>

<style scoped>
.website-select {
    width: min(18rem, 100%);
}
</style>
